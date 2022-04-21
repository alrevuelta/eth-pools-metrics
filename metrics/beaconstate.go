package metrics

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	//"github.com/alrevuelta/eth-pools-metrics/prometheus"
	"github.com/alrevuelta/eth-pools-metrics/pools"
	"github.com/alrevuelta/eth-pools-metrics/postgresql"
	"github.com/alrevuelta/eth-pools-metrics/prometheus"
	"github.com/alrevuelta/eth-pools-metrics/schemas"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/rs/zerolog"

	log "github.com/sirupsen/logrus"
)

type BeaconState struct {
	httpClient    *http.Service
	eth1Endpoint  string
	eth2Endpoint  string
	pg            *postgresql.Postgresql
	fromAddresses []string
	poolNames     []string
}

func NewBeaconState(
	eth1Endpoint string,
	eth2Endpoint string,
	pg *postgresql.Postgresql,
	fromAddresses []string,
	poolNames []string) (*BeaconState, error) {

	client, err := http.New(context.Background(),
		http.WithTimeout(60*time.Second),
		http.WithAddress(eth2Endpoint),
		http.WithLogLevel(zerolog.WarnLevel),
	)
	if err != nil {
		return nil, err
	}

	httpClient := client.(*http.Service)

	return &BeaconState{
		httpClient:    httpClient,
		eth2Endpoint:  eth2Endpoint,
		pg:            pg,
		fromAddresses: fromAddresses,
		poolNames:     poolNames,
		eth1Endpoint:  eth1Endpoint,
	}, nil
}

func (p *BeaconState) Run() {
	var prevEpoch uint64 = uint64(0)
	var prevBeaconState *spec.VersionedBeaconState = nil

	for {
		// Before doing anything, check if we are in the next epoch
		headSlot, err := p.httpClient.NodeSyncing(context.Background())
		if err != nil {
			log.Error("Could not get node sync status:", err)
			continue
		}

		if headSlot.IsSyncing {
			log.Error("Node is not in sync")
			continue
		}

		// The closer to head, the more risk of having inaccurate data due
		// to being in a fork. Metrics could also be fetched on finalized epochs.
		// TODO: Don't hardcode 32
		currentEpoch := uint64(headSlot.HeadSlot)/uint64(32) - 3

		if prevEpoch >= currentEpoch {
			// do nothing
			time.Sleep(5 * time.Second)
			continue
		}

		// TODO: Retry once if fails
		currentBeaconState, err := p.GetBeaconState(currentEpoch)
		if err != nil {
			prevBeaconState = nil
			log.Error("Error fetching beacon state:", err)
			continue
		}

		// if no prev beacon state is known, fetch it
		if prevBeaconState == nil {
			prevBeaconState, err = p.GetBeaconState(currentEpoch - 1)
			// TODO: Retry
			if err != nil {
				log.Error(err)
				continue
			}
		}

		// General network metrics
		nOfSlashedValidators := 0
		validators := GetValidators(currentBeaconState)

		// Create a map to convert from key to index for quick access
		valKeyToIndex := PopulateKeysToIndexesMap(currentBeaconState)

		for _, val := range validators {
			if val.Slashed {
				nOfSlashedValidators++
			}
		}
		prometheus.TotalDepositedValidators.Set(float64(len(validators)))
		prometheus.TotalSlashedValidators.Set(float64(nOfSlashedValidators))
		log.WithFields(log.Fields{
			"Total Validators":         len(validators),
			"Total Slashed Validators": nOfSlashedValidators,
		}).Info("Network stats:")

		// TODO: Move somewhere else
		syncCommitteeKeys := BLSPubKeyToByte(GetCurrentSyncCommittee(currentBeaconState))
		syncCommitteeIndexes := GetIndexesFromKeys(syncCommitteeKeys, valKeyToIndex)

		for _, poolName := range p.poolNames {
			var pubKeysDeposited [][]byte

			// Special case: hardcoded keys
			if strings.HasSuffix(poolName, ".txt") {
				pubKeysDeposited, err = pools.ReadCustomValidatorsFile(poolName)
				if err != nil {
					log.Fatal(err)
				}
				// trim the file path and extension
				poolName = filepath.Base(poolName)
				poolName = strings.TrimSuffix(poolName, filepath.Ext(poolName))
			} else if poolName == "rocketpool" {
				pubKeysDeposited = pools.RocketPoolKeys
				// From known from-addresses
			} else {
				poolAddressList := pools.PoolsAddresses[poolName]
				log.Info("The pool:", poolName, " from-address are: ", poolAddressList)
				pubKeysDeposited, err = p.pg.GetKeysByFromAddresses(poolAddressList)
				if err != nil {
					log.Error(err)
					continue
				}
			}

			log.Info("The pool:", poolName, " contains ", len(pubKeysDeposited), " keys (may be hardcoded)")

			if len(pubKeysDeposited) == 0 {
				log.Warn("No deposited keys for: ", poolName, ", skipping")
				continue
			}

			validatorIndexes := GetIndexesFromKeys(pubKeysDeposited, valKeyToIndex)
			activeValidatorIndexes := GetActiveIndexes(validatorIndexes, currentBeaconState)

			log.Info("The pool:", poolName, " contains ", len(validatorIndexes), " validators detected in the beacon state")
			log.Info("The pool:", poolName, " contains ", len(activeValidatorIndexes), " active validators detected in the beacon state")
			//log.Info(validatorIndexes)

			metrics, err := PopulateParticipationAndBalance(
				activeValidatorIndexes,
				currentBeaconState,
				prevBeaconState)

			if err != nil {
				log.Error(err)
				continue
			}

			// TODO: Move somewhere else
			// Sync committee information. For each pool, shows the validators part of the committee
			poolSyncIndexes := GetValidatorsIn(syncCommitteeIndexes, activeValidatorIndexes)

			log.Info("Pool: ", poolName, " sync committee validators ", poolSyncIndexes)

			logMetrics(metrics, poolName)
			setPrometheusMetrics(metrics, poolSyncIndexes, poolName)
		}

		prevBeaconState = currentBeaconState
		prevEpoch = currentEpoch
	}
}

func GetValidatorsIn(allSyncCommitteeIndexes []uint64, poolValidatorIndexes []uint64) []uint64 {
	poolCommmitteeIndexes := make([]uint64, 0)
	for i := range allSyncCommitteeIndexes {
		for j := range poolValidatorIndexes {
			if allSyncCommitteeIndexes[i] == poolValidatorIndexes[j] {
				poolCommmitteeIndexes = append(poolCommmitteeIndexes, allSyncCommitteeIndexes[i])
				break
			}
		}
	}
	return poolCommmitteeIndexes
}

func PopulateKeysToIndexesMap(beaconState *spec.VersionedBeaconState) map[string]uint64 {
	// TODO: Naive approach. Reset the map every time
	valKeyToIndex := make(map[string]uint64, 0)
	for index, beaconStateKey := range GetValidators(beaconState) {
		valKeyToIndex[hex.EncodeToString(beaconStateKey.PublicKey[:])] = uint64(index)
	}
	return valKeyToIndex
}

// TODO: Move to utils
func BLSPubKeyToByte(blsKeys []phase0.BLSPubKey) [][]byte {
	keys := make([][]byte, 0)
	for i := range blsKeys {
		keys = append(keys, blsKeys[i][:])
	}
	return keys
}

// Make sure the validator indexes are active
func PopulateParticipationAndBalance(
	activeValidatorIndexes []uint64,
	beaconState *spec.VersionedBeaconState,
	prevBeaconState *spec.VersionedBeaconState) (schemas.ValidatorPerformanceMetrics, error) {

	metrics := schemas.ValidatorPerformanceMetrics{
		EarnedBalance:    big.NewInt(0),
		LosedBalance:     big.NewInt(0),
		TotalBalance:     big.NewInt(0),
		EffectiveBalance: big.NewInt(0),
		TotalRewards:     big.NewInt(0),
	}

	nOfIncorrectSource, nOfIncorrectTarget, nOfIncorrectHead, indexesMissedAtt := GetParticipation(
		activeValidatorIndexes,
		beaconState)

	currentBalance, currentEffectiveBalance := GetTotalBalanceAndEffective(activeValidatorIndexes, beaconState)
	prevBalance, prevEffectiveBalance := GetTotalBalanceAndEffective(activeValidatorIndexes, prevBeaconState)

	// Make sure we are comparing apples to apples
	if currentEffectiveBalance.Cmp(prevEffectiveBalance) != 0 {
		return schemas.ValidatorPerformanceMetrics{},
			errors.New(fmt.Sprint("Can't calculate delta balances, effective balances are different:",
				currentEffectiveBalance, " vs ", prevEffectiveBalance))
	}

	rewards := big.NewInt(0).Sub(currentBalance, currentEffectiveBalance)
	deltaEpochBalance := big.NewInt(0).Sub(currentBalance, prevBalance)

	lessBalanceIndexes, earnedBalance, lostBalance, err := GetValidatorsWithLessBalance(
		activeValidatorIndexes,
		prevBeaconState,
		beaconState)

	if err != nil {
		return schemas.ValidatorPerformanceMetrics{}, err
	}

	metrics.IndexesLessBalance = lessBalanceIndexes
	metrics.EarnedBalance = earnedBalance
	metrics.LosedBalance = lostBalance

	// TODO: Don't hardcode 32
	metrics.Epoch = GetSlot(beaconState) / 32

	metrics.NOfTotalVotes = uint64(len(activeValidatorIndexes)) * 3
	metrics.NOfIncorrectSource = nOfIncorrectSource
	metrics.NOfIncorrectTarget = nOfIncorrectTarget
	metrics.NOfIncorrectHead = nOfIncorrectHead
	metrics.NOfValidatingKeys = uint64(len(activeValidatorIndexes))
	//metrics.NOfValsWithLessBalance = nOfValsWithDecreasedBalance
	//metrics.EarnedBalance = earned
	//metrics.LosedBalance = losed
	metrics.IndexesMissedAtt = indexesMissedAtt
	//metrics.LostBalanceKeys = lostKeys
	metrics.TotalBalance = currentBalance
	metrics.EffectiveBalance = currentEffectiveBalance
	metrics.TotalRewards = rewards
	metrics.DeltaEpochBalance = deltaEpochBalance

	return metrics, nil
}

// TODO: Get slashed validators

func (p *BeaconState) GetBeaconState(epoch uint64) (*spec.VersionedBeaconState, error) {
	log.Info("Fetching beacon state for epoch: ", epoch)
	slotStr := strconv.FormatUint(epoch*32, 10)
	beaconState, err := p.httpClient.BeaconState(
		context.Background(),
		slotStr)
	if err != nil {
		return nil, err
	}
	log.Info("Got beacon state for epoch:", GetSlot(beaconState)/32)
	return beaconState, nil
}

func GetTotalBalanceAndEffective(
	activeValidatorIndexes []uint64,
	beaconState *spec.VersionedBeaconState) (*big.Int, *big.Int) {

	totalBalances := big.NewInt(0).SetUint64(0)
	effectiveBalance := big.NewInt(0).SetUint64(0)
	validators := GetValidators(beaconState)
	balances := GetBalances(beaconState)

	for _, valIdx := range activeValidatorIndexes {
		// Skip if index is not present in the beacon state
		if valIdx >= uint64(len(balances)) {
			log.Warn("validator index goes beyond the beacon state indexes")
			continue
		}
		valBalance := big.NewInt(0).SetUint64(balances[valIdx])
		//log.Info(valIdx, ":", valBalance)
		valEffBalance := big.NewInt(0).SetUint64(uint64(validators[valIdx].EffectiveBalance))
		totalBalances.Add(totalBalances, valBalance)
		effectiveBalance.Add(effectiveBalance, valEffBalance)
	}
	return totalBalances, effectiveBalance
}

// Returns the indexes of the validator keys. Note that the indexes
// may belong to active, inactive or even slashed keys.
func GetIndexesFromKeys(
	validatorKeys [][]byte,
	valKeyToIndex map[string]uint64) []uint64 {

	indexes := make([]uint64, 0)

	// Use global prepopulated map
	for _, key := range validatorKeys {
		if valIndex, ok := valKeyToIndex[hex.EncodeToString(key)]; ok {
			indexes = append(indexes, valIndex)
		} else {
			log.Warn("Index for key: ", hex.EncodeToString(key), " not found in beacon state")
		}
	}

	return indexes
}

func GetActiveIndexes(
	validatorIndexes []uint64,
	beaconState *spec.VersionedBeaconState) []uint64 {

	activeIndexes := make([]uint64, 0)

	validators := GetValidators(beaconState)
	beaconStateEpoch := GetSlot(beaconState) / 32

	for _, valIdx := range validatorIndexes {
		if beaconStateEpoch >= uint64(validators[valIdx].ActivationEpoch) {
			activeIndexes = append(activeIndexes, valIdx)
		}
	}

	return activeIndexes
}

func GetValidatorsWithLessBalance(
	activeValidatorIndexes []uint64,
	prevBeaconState *spec.VersionedBeaconState,
	currentBeaconState *spec.VersionedBeaconState) ([]uint64, *big.Int, *big.Int, error) {

	prevEpoch := GetSlot(prevBeaconState) / 32
	currEpoch := GetSlot(currentBeaconState) / 32
	prevBalances := GetBalances(prevBeaconState)
	currBalances := GetBalances(currentBeaconState)

	if (prevEpoch + 1) != currEpoch {
		return nil, nil, nil, errors.New(fmt.Sprintf(
			"epochs are not consecutive: slot %d vs %d",
			prevEpoch,
			currEpoch))
	}

	indexesWithLessBalance := make([]uint64, 0)
	earnedBalance := big.NewInt(0)
	lostBalance := big.NewInt(0)

	for _, valIdx := range activeValidatorIndexes {
		// handle if there was a new validator index not register in the prev state
		if valIdx >= uint64(len(prevBalances)) {
			log.Warn("validator index goes beyond the beacon state indexes")
			continue
		}

		prevEpochValBalance := big.NewInt(0).SetUint64(prevBalances[valIdx])
		currentEpochValBalance := big.NewInt(0).SetUint64(currBalances[valIdx])
		delta := big.NewInt(0).Sub(currentEpochValBalance, prevEpochValBalance)

		if delta.Cmp(big.NewInt(0)) == -1 {
			indexesWithLessBalance = append(indexesWithLessBalance, valIdx)
			lostBalance.Add(lostBalance, delta)
		} else {
			earnedBalance.Add(earnedBalance, delta)
		}
	}

	return indexesWithLessBalance, earnedBalance, lostBalance, nil
}

// See spec: from LSB to MSB: source, target, head.
// https://github.com/ethereum/consensus-specs/blob/master/specs/altair/beacon-chain.md#participation-flag-indices
func GetParticipation(
	activeValidatorIndexes []uint64,
	beaconState *spec.VersionedBeaconState) (uint64, uint64, uint64, []uint64) {

	indexesMissedAtt := make([]uint64, 0)

	validators := GetValidators(beaconState)
	previousEpochParticipation := GetPreviousEpochParticipation(beaconState)

	var nIncorrectSource, nIncorrectTarget, nIncorrectHead uint64

	for _, valIndx := range activeValidatorIndexes {
		// Ignore slashed validators
		if validators[valIndx].Slashed {
			continue
		}
		beaconStateEpoch := GetSlot(beaconState) / 32
		// Ignore not yet active validators
		// TODO: Test this
		if uint64(validators[valIndx].ActivationEpoch) > beaconStateEpoch {
			//log.Warn("index: ", valIndx, " is not active yet")
			continue
		}

		// TODO: Dont know why but Infura returns 0 for all CurrentEpochAttestations
		epochAttestations := previousEpochParticipation[valIndx]
		if !isBitSet(uint8(epochAttestations), 0) {
			nIncorrectSource++
			indexesMissedAtt = append(indexesMissedAtt, valIndx)
		}
		if !isBitSet(uint8(epochAttestations), 1) {
			nIncorrectTarget++
		}
		if !isBitSet(uint8(epochAttestations), 2) {
			nIncorrectHead++
		}
	}
	return nIncorrectSource, nIncorrectTarget, nIncorrectHead, indexesMissedAtt
}

/* TODO: Unused. Add support for Bellatrix
func GetInactivityScores(
	activeValidatorIndexes []uint64,
	beaconState *spec.VersionedBeaconState) []uint64 {
	inactivityScores := make([]uint64, 0)
	for _, valIdx := range activeValidatorIndexes {
		inactivityScores = append(inactivityScores, beaconState.Altair.InactivityScores[valIdx])
	}
	return inactivityScores
}*/

// Check if bit n (0..7) is set where 0 is the LSB in little endian
func isBitSet(input uint8, n int) bool {
	return (input & (1 << n)) > uint8(0)
}

func logMetrics(
	metrics schemas.ValidatorPerformanceMetrics,
	poolName string) {
	balanceDecreasedPercent := (float64(len(metrics.IndexesLessBalance)) / float64(metrics.NOfValidatingKeys)) * 100

	log.WithFields(log.Fields{
		"PoolName":                    poolName,
		"Epoch":                       metrics.Epoch,
		"nOfTotalVotes":               metrics.NOfTotalVotes,
		"nOfIncorrectSource":          metrics.NOfIncorrectSource,
		"nOfIncorrectTarget":          metrics.NOfIncorrectTarget,
		"nOfIncorrectHead":            metrics.NOfIncorrectHead,
		"nOfValidators":               metrics.NOfValidatingKeys,
		"PercentIncorrectSource":      (float64(metrics.NOfIncorrectSource) / float64(metrics.NOfTotalVotes)) * 100,
		"PercentIncorrectTarget":      (float64(metrics.NOfIncorrectTarget) / float64(metrics.NOfTotalVotes)) * 100,
		"PercentIncorrectHead":        (float64(metrics.NOfIncorrectHead) / float64(metrics.NOfTotalVotes)) * 100,
		"nOfValsWithDecreasedBalance": len(metrics.IndexesLessBalance),
		"balanceDecreasedPercent":     balanceDecreasedPercent,
		"epochEarnedBalance":          metrics.EarnedBalance,
		"epochLostBalance":            metrics.LosedBalance,
		"totalBalance":                metrics.TotalBalance,
		"effectiveBalance":            metrics.EffectiveBalance,
		"totalRewards":                metrics.TotalRewards,
		"ValidadorKeyMissedAtt":       metrics.IndexesMissedAtt,
		"ValidadorKeyLessBalance":     metrics.IndexesLessBalance,
		"DeltaEpochBalance":           metrics.DeltaEpochBalance,
	}).Info(poolName + " Stats:")
}

func setPrometheusMetrics(
	metrics schemas.ValidatorPerformanceMetrics,
	numSyncValidators []uint64,
	poolName string) {

	prometheus.TotalBalanceMetrics.WithLabelValues(
		poolName).Set(float64(metrics.TotalBalance.Int64()))

	prometheus.ActiveValidatorsMetrics.WithLabelValues(
		poolName).Set(float64(metrics.NOfValidatingKeys))

	prometheus.IncorrectSourceMetrics.WithLabelValues(
		poolName).Set(float64(metrics.NOfIncorrectSource))

	prometheus.IncorrectTargetMetrics.WithLabelValues(
		poolName).Set(float64(metrics.NOfIncorrectTarget))

	prometheus.IncorrectHeadMetrics.WithLabelValues(
		poolName).Set(float64(metrics.NOfIncorrectHead))

	prometheus.EpochEarnedAmountMetrics.WithLabelValues(
		poolName).Set(float64(metrics.EarnedBalance.Int64()))

	prometheus.EpochLostAmountMetrics.WithLabelValues(
		poolName).Set(float64(metrics.LosedBalance.Int64()))

	prometheus.DeltaEpochBalanceMetrics.WithLabelValues(
		poolName).Set(float64(metrics.DeltaEpochBalance.Int64()))

	prometheus.CumulativeConsensusRewards.WithLabelValues(
		poolName).Set(float64(metrics.TotalRewards.Int64()))

	prometheus.NumOfSyncCommitteeValidators.WithLabelValues(
		poolName).Set(float64(len(numSyncValidators)))

	// TODO: Add the indexes of the sync committees
	// TODO: Add if the sync committees are fulfilling their duties or not

	// TODO: Remove this from here and from prometheus
	//prometheus.NOfTotalVotes.Set(float64(metrics.NOfTotalVotes))
	//prometheus.NOfIncorrectSource.Set(float64(metrics.NOfIncorrectSource))
	//prometheus.NOfIncorrectTarget.Set(float64(metrics.NOfIncorrectTarget))
	//prometheus.NOfIncorrectHead.Set(float64(metrics.NOfIncorrectHead))
	//prometheus.EarnedAmountInEpoch.Set(float64(metrics.EarnedBalance.Int64()))
	//prometheus.LosedAmountInEpoch.Set(float64(metrics.LosedBalance.Int64()))
	//prometheus.CumulativeRewards.Set(float64(metrics.TotalRewards.Int64()))
	//prometheus.TotalBalance.Set(float64(metrics.TotalBalance.Int64()))
	//prometheus.EffectiveBalance.Set(float64(metrics.EffectiveBalance.Int64()))

	// TODO: Deprecate this, send the raw number
	balanceDecreasedPercent := (float64(metrics.NOfValsWithLessBalance) / float64(metrics.NOfValidatingKeys)) * 100
	prometheus.BalanceDecreasedPercent.Set(balanceDecreasedPercent)

	/* TODO:
	for _, v := range metrics.MissedAttestationsKeys {
		prometheus.MissedAttestationsKeys.WithLabelValues(v).Inc()
	}

	for _, v := range metrics.LostBalanceKeys {
		prometheus.LessBalanceKeys.WithLabelValues(v).Inc()
	}*/
}

// Wrappers on top of the beacon state to fetch some fields regardless of Altair or Bellatrix
// Note that this is needed because both block types do not implement the same interface, since
// the state differs accross versions.
// Note also that this functions only make sense for the beacon state fields that are common
// to all the versioned beacon states.
func GetValidators(beaconState *spec.VersionedBeaconState) []*phase0.Validator {
	var validators []*phase0.Validator
	if beaconState.Altair != nil {
		validators = beaconState.Altair.Validators
	} else if beaconState.Bellatrix != nil {
		validators = beaconState.Bellatrix.Validators
	} else {
		log.Fatal("Beacon state was empty")
	}
	return validators
}

func GetBalances(beaconState *spec.VersionedBeaconState) []uint64 {
	var balances []uint64
	if beaconState.Altair != nil {
		balances = beaconState.Altair.Balances
	} else if beaconState.Bellatrix != nil {
		balances = beaconState.Bellatrix.Balances
	} else {
		log.Fatal("Beacon state was empty")
	}
	return balances
}

func GetPreviousEpochParticipation(beaconState *spec.VersionedBeaconState) []altair.ParticipationFlags {
	var previousEpochParticipation []altair.ParticipationFlags
	if beaconState.Altair != nil {
		previousEpochParticipation = beaconState.Altair.PreviousEpochParticipation
	} else if beaconState.Bellatrix != nil {
		previousEpochParticipation = beaconState.Bellatrix.PreviousEpochParticipation
	} else {
		log.Fatal("Beacon state was empty")
	}
	return previousEpochParticipation
}

func GetSlot(beaconState *spec.VersionedBeaconState) uint64 {
	var slot uint64
	if beaconState.Altair != nil {
		slot = beaconState.Altair.Slot
	} else if beaconState.Bellatrix != nil {
		slot = beaconState.Bellatrix.Slot
	} else {
		log.Fatal("Beacon state was empty")
	}
	return slot
}

func GetCurrentSyncCommittee(beaconState *spec.VersionedBeaconState) []phase0.BLSPubKey {
	var pubKeys []phase0.BLSPubKey
	if beaconState.Altair != nil {
		pubKeys = beaconState.Altair.CurrentSyncCommittee.Pubkeys
	} else if beaconState.Bellatrix != nil {
		pubKeys = beaconState.Bellatrix.CurrentSyncCommittee.Pubkeys
	} else {
		log.Fatal("Beacon state was empty")
	}
	return pubKeys
}
