package pools

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	log "github.com/sirupsen/logrus"
)

var consensuscodefiKeysStr = []string{
	"0x8a54ad44fba9bf455db21a806b7019473ebdc2e18408c31c81ccc0b246227f0033ec08bd96086ea973073a40af4f837d",
	"0x9890923406d257999bf433959008f15cc31f50dfe8318717615421eba9e7459d7bbebb612e33f5e8fe9fca2fcd1a79d9",
	"0xadb95eeac8d0163faf117fc6e54113df5c8e3ea831a123560849939117d7ab96fc9dbc49ec41109db8deb7add47f32f4",
	"0xa9260979a22027eaa2cc71e86e5942f8e6edcd0dd9cd391132b2dff5f6b664de697c4c79bbc6fa8f820140093dfdfe76",
	"0xa43a0869eb2472cc9cdd80379751e080a0854229b149a8f15e1eb03dbf92abf5c66dda1e9d10bf59aa68772ef36e7450",
	"0xa95a8e556e1c1ffaad910ca771921027557ebe8c413e33abe3e0884b6b41ecce42493a7744e48a2e9e684d3cef8b8585",
	"0x95106d50db100f5497059884bcafb5ed4e812c893922044df09b4a3bc4b60e48f01f3eb5ef52aed708c7863b85984042",
	"0x90fd7cc0264b7b3fda121da454f61962b718ac1b2a28f75d1d8ab3beeb4ec321df4682aac6f1a8e3925773520085e2f1",
	"0xa04172193d1a884dfa65654cc5a32d11838e8c31f3f233308a64be87efc28b963d67c40782656e4b234c3e24fd0123da",
	"0x9974cfc5cb42919b283c3a944ac3b506c86eca144e2799449e3f43fc6f7f6784cc1af8cf15440471bd2f74c9e0d2602a",
	"0xa5d49328c80b0a8ef1d5a7bfb289481892d2c8d073b82771c641724696044bd51e693bea99b7db94e7feb0ca6c427295",
	"0xaa1dc67bd70109d319bb3c93d3a3a4547be9887b9d9289fb10c593a908d939f1162b18a5ca6064f2fab04c1a04ddfd27",
	"0xa1acfa8ad1978bf506253ed58cc6874e60709d849a48470423701dad008136cc5206b36a044c8d55df2da1b8239f16d6",
	"0xac39bb90bb48f9a3fede6c1ba062eb025554dc6f5df747ce17dbfd7d30839877750f154a3ae0465914cedf5008b29560",
	"0xab0220a3a5edb5ace0e8d3a5f4c41b80d2d2975656c67e72f5ea653af4225f765c3f627226147d6808d71d1f006baa14",
	"0xa7806f192a9e83e1a861076218528e58f0aaea22430c50bd5cfe54812ed3529ecd44026afc0c09e8dda5e69d73f42fda",
	"0x88985fe23d4f3aa8c13ca171237268dc7e711c7f475ebe68c7a647df7cd772ed8375ab3152f5b05cebcb4dbeef056d43",
	"0x840e00d2cb012ca892ab4a19e97bc22d1e446108e69c830d9dfeaf28d206144362ba67309da0f1e9868b608f74c029cb",
	"0x90ed2983c96a14e4c5cb5fb9fff86e0ba07db1d39225713ef026d9ea702bd362d84969d6886c6d88b0d8ba5e9244d373",
	"0x8d6f1e449bac3f41224efff7988749e224f7503c5a068e8b23a2de9d7b21ceba15406e76324e163385638928b0803276",
	"0x99b6eecc20f9771e30c63b08c79d2d371435ae236437a425438792209850a9ff0e29d74d4bb4bd7e1463d2431b655d92",
	"0xa2dea8827c5a67c72d10188beccc922060181a188572a1b2ef8032e7efc7e09381cf9b88bfe8a446e98b5c7d14458fa6",
	"0xa5f45270854e9af66d8fed6c0343b1be48937c9b2407baba05ed05cbe8f8c738e72e0022c7984cf63c0d945bc35c305d",
	"0xb66091eb3ee860bc082cbffa3f0ee3ac23e184ce21249efed312ac932838e135e6fcc19b8bb93914543a51275ce2119c",
	"0x825f2f8922583dd1f111d3fd53b3025c95ab48213921f916c3a854fb90b6579991c5163bac950bc08c99e4aee6202205",
	"0x8beb0d47bbdf3df4da909d211b6b1a54404b700dcc793a14632e3eeb12222e4ab7dbb3ada4f15d88788278ced868cac2",
	"0xaf1b0f2d0d327297c4f9700b2dc24b2f7c591b78895d570b359ab162197ddbdb3560506a8517e67c5769c661bc3ec8d0",
	"0x82c3cc27a87a36ccd1c5be12cfa250d0a852fc14c7121eaf215811c0595e97ca307769d48e41675d1972adbd54e559e2",
	"0x904267e3a5cbd83486bb974190d1b723c4a0bbd6240026fe351470a1ab1d3b47b9815d0fa85bb7c2effd2e7d6c95cb50",
	"0x831e559b8bbb5b802c9d7fbae8760df771f58ca3dc5ce95b61de0b03a0290c64cc79c2b329addc087ab7b53acc40f840",
	"0x9680a09b1f991dab97da9766488eb2968994a313ebbac540fc6f175f591bd92727bb22f4cb582d95723d469c0d47a428",
	"0xad0a7f3a1c5c5f482f1d4ffa8cafbef21ac5a4d8c5c300469c30f1199a715e9cad1fe7ec7f440d1f4aada708e89dfd69",
	"0x940f22ad2caed930de0fae3769f5d7b3ea5dfcf0303d0dad3e5a977ebc34f1fe99f11259ad5b4823f09b5487821724a3",
	"0xb16ee25b5302e3e9012ae145b1251f0e0c79256a58975303d0496ec52f7e33b8dd38d883ce9292fd283424e6910ac6b9",
	"0xb1f089bc1160bba92b19054d4bd0f7efdbf2beecd7f1542388b7fe3048287a4364739bd47d4d0901002529444a6ad91a",
	"0xb8e8318688889bda888c6beef0cef634a9a0127cc3ee4ae96522716708a189fbe1f39fa84493baf669bb95d95ae4e4a3",
	"0xa1b9a3373af79cfd780d5f8a72be4821c7516517ad2c7d13946011c9a411c78e43369a6f13496f35192369e47b076c42",
	"0xa08f007a8d9048b61045f9ffb71b60227206ef5b63f5d02ef342a42682b90c1b78a700e1998bee8f1fdb86fce8ffde81",
	"0xb118b5ff0a1b488fdfa26a4699ca5e272e9815c8e8f834de7b29c2bedf9d5909fd781fd873f63306ce2b807ceb676286",
	"0xaecfc0c5c3d14b8b67e7318c7a0a50882637909fa4b1249758ae3d4d22f371b328e1086aa7c308c295f723efc8b96eb2",
	"0xa76a3f38c2db2bea23d4bafe211bb1deb88811e6dd3fd3e1d2d9e09fb948752783843fb9519c05ce561127048c6bf75c",
	"0x9429a6637fd8216ddcc0598c9f8dc76c1dcc3fbb96f94d4b3357e93d43d4fdb211f544e8a21b3bc8ee25fcadebe05e3f",
	"0x8d3ee750d2469492b443857c6155c590b5fa5773ee64444d7cb75e2130b1c593d8cd5fb42ea90d4698e7e5af728c15bd",
	"0x90aea76203380c1a8b8f3564ef8288eeef9753a5cf9216f7e6d0ba5b0df5bf3d70837fbc0b33e2f3e2008bfc9aa4af00",
	"0x895b1b978869db8ea882c2440063384602b92f4d8d5c39278a5cb1f0cd49339d0cb08de408240a0bbbe8f29f47013a8b",
	"0x80c48ef3acadf6c207d1093a503003135cf33283acc097df9b79d46b4646db9826e0c663a94adf7eaa32471e386689ba",
	"0xa0cb0cfff680de94b86aef0b85a0b3d4acb174a93a6b8a1a7f793af4cb0621b2f5e29b2dcce40ebd743db0b41455e604",
	"0xaad06f5ef8a1d3deaeda049905984707ecb6a7b37d72c308f1b67307bf61b2b20569f357b86a938c650491f8c86dfc9f",
	"0x815c1b15b3047b15eb516c96d34c17fc0c6156a6bbd62bdad3cda1bff6a7e2f8324322576b3c686777ce1e6877c8d1d2",
	"0xadaa31cb03a38f1b4669c8d209714f03e5f732b932feda19e410c16997405edafe12d24f10b2630b1cf4cdc752414589",
	"0xaed1751e309a6002dd0af10cc437f83eb6b01ae43982e455ca1f294f0b931478c9d723168bf2570ff77e255b1ad06b2d",
	"0x981b228789cb986f21518ca778341cba62e393a98b1822a892f5fe3fb23f74fe30fbbc4093490809feeca83fc7e5ebbf",
	"0xb416b9ed6a1b8c7fd7b1afec47121eabff48238dc49ddef4c39994701f97ce8e041abdb5d9ffc9808779b4cc8432589e",
	"0x95b37cacde945712e5fc0c7946ee4f28b32106fb9a7e7fc37a95d937922b41bf6ce0d677f9cb0a965ed21193c3e5a0af",
	"0x90a1d2e3df2c6c3f98aab2f15a40c11b6d1a0046b45c1c8e68ddc0fe035d98ad1ff98302229f81f1483f988f73bd866e",
	"0x8a795e153dee54e6315cd4285c04806c92682f9c3b1651893bb2c6d5a3ec1f575593453e045c9a3c41883c47e271560d",
	"0x86ab87391a154dccd046a24f9e82a750c838e0024a36d2c6567a9654fd981a0c59538fa9475b48ba5e6e9ae1eba81b60",
	"0xa600672829dabf9e159f0582cf211d515a74fc154a08f00ff37669afa17d5855b8fccf0421a7e5b7cc988ce369805429",
	"0xa927de934782f586e3455f87759bd3a07c022fb3220b9c950d4741a30a763e68db93c9defa4f02113bfa3ad822fa36e3",
	"0x94e739cbfb5627d6be3e72ad654e0d451d8415dee0ffa141ce2098c65553fa4c7449d74c549b2d988a7274f23829b454",
	"0x8091fd258234b5bfb8c964877ea7822b2cb18f0a12b171ed3c9d1b65a9a5caa57ffe2584fe35987459cd2e50b0c53c70",
	"0xb29330cf86d87a96ef37acec4530700b1952c1bc2d6bfce82eeec7bbfcb053df844edc8b53018bdcdc8d3e6022073bdf",
	"0x9229007096359834632d188875ea02b2b65ec67a6af1eb8a70312bb2eb3b537f0b427281a0a9f438adca41b314ee4ce8",
	"0xaee7d69b71660c3bc6af35f76a8cc2b2f3a76af8965a352adff901e756459a06f2cc908c89d50fb49b836f8d44104681",
	"0xa607607ef39644d66a5282895892c62af5c975c1635ba023bfd0bd24d0013d6504a18573db060f35d2ab97eca5bfd627",
	"0x8f638fb564732cd766b200ef5c70b7e25857aa415501fdc55103b34a80a5f71a807d25ad1a4e445a41897066469470c5",
	"0x834f7e8f248ac6578840e7c2d9b2d8a79d2695f405bd9b0d8465e943e5ffe284aac41e990e5b304475c755fef2d99b7b",
	"0x84f5aa2a8af494675f1c442ca15bc89d14a4a68a845d82bf8bbc8b0984290c9608984cb8bc580b968412cb15419863eb",
	"0x97f138abd60c1cead35916bf2561763d11d6283da465e3010cf02bb5811592cd3ffd2fc68a3034051826a6b526f86531",
	"0xb9a922868cdb50803b935419354d6a6438857024376f3d680ec085ff269b5715806fbb8d4ed5344cfd10f90702165370",
	"0x94ae313845f8e5392d4d6b04c66b9a0a482dee3358d9af2df697f332c1c785371399013fcbfc7a9e7ac20f2eaf887675",
	"0xa8582384b8ba65fc76ea792a703a0731ce3d1c2d786e7c6778a59d499e6e95c6394b582788681a6fa1cabb0bc3a4bb4f",
	"0xaf13cfa25ed26f27ca9295adcb56cf8dbbffe37126ec48509bd00bd9cbcbe54494602ebc7c894884e5532464b80ca499",
	"0xb6b592a7fa07ae89a3f8b93522f5cc05a5acb2735146ee1716445b8ef1b1ce4e8e8c2de613bc4dab0daa9e498f6a344d",
	"0xb69ff7f54afbc521a3369613423e8c25b4e9c9f7a8a4fa2a88c836d66f6b8d662ca0333d35a3a6f8d38d9fa57c3efd1a",
	"0xa35a69d381f5729707298aa828a2a0fa23f7433d30bb9e67f2cf627a6e575cdbd1339e8289e0828a6e0d37c159eedd87",
	"0x8998ad86994d975d17ed3f9e758099df10e762de878efb92af269a07ba71bc9aa899c41a8b6e924dd7a9b666f8f08a28",
	"0x97e4c5a4669f24ca26bf11970a73e001d3e6e0fe3c19e908e5ad8837de1e2cfa5ca0bffbe677b99fd8738fa063cd2356",
	"0xa37a2936a34d5465090b154d9f5f8482eab57f33e88a52c0f9e73a4185a73ecce1af4fc183157bb23687df55de053828",
	"0xac858205f5f46c240f60c0b8fc5a2a2aef9b915ba086075536f92ae8bd1228b0f9927f0e94a1f731c6f675d3ec3ff330",
	"0x93131002dbda37f346ff4fe7eeed2636a5db45a0cee1f2b83f3aeb185fe025b9c77d387ce0b21098c1fde154d8e66e93",
	"0x92a92af3a7ac894da9ca9f252d7a861bf44c3a30ed8db1c9e58a7ed764fe0abb0c49ddd11fa1401e1edf860e445800a8",
	"0x82cf0d32c985ff0157cb53711ea116cedd43fcaa923fe69bb15ad365ffabda86d7629aa4e7ace788d9b171aa6fe20f8b",
	"0xa5bdcd73c24cfdee5ecd8fc889ed2c51441d78a7d01277753d956899588442060e0d99fedb5e8a44becc05f3ffd81d76",
	"0x989ea5acf4da944ee9c1e9e2a71bb88e4f49a843f3d40f1b34b1a0c5c57bcff15321a4aa1e70d1f0b3f54ff9fb382c07",
	"0xa2fade8ce30eaace52a318d393ae9c63bd79452fd7be0f01c3fecd4927c89e1a4b779f4a7add5c2437c68f3fa6e8cdee",
	"0x847e4971953737f99dd6176710597cd8bb33c3c9c1e2e8ea7118bfb02e4ed1ec327a651355e92175d007cc837373d9ee",
	"0x8ab21c2cd81bd2d90c48528005992386458429129b1185ac220cadef85055d6d89cd7097980742b45ea1e3d5c3bad91f",
	"0x8458d1ffadac35f4e14fedf9c3c3dc8d2a76f1067ffb2624b8074f523424387a97651dcb09bbe96f0ca886daaf8f2066",
	"0xa11fa5329331cd24e51c62226313bc475211ffcdb7de8dd9796e16bf15503222f86108d48cca32467ad66590f690f469",
	"0xa9431cd1fcf731c6dc448a0f36014652cbd521153a0edaf1ea4787c9759bd0fb56df2f0f77ce5dc99a2fe5eebfd96e63",
	"0xa092cccc8d0685dbc63707c731004320713e900093f64a70d19fab998f9b1b989023e640d4e0e3f82957fdead193c6f0",
	"0xb5470bee0b2a749c881178071c03e8d28316e2256f0d9532ccce5f2d438ee20cb7046feb3ccc4463ec37155b6dd363bb",
	"0x973dd85d4df618722443bd08fdbda9c870558313d10f26144b300067ba99142f493c38015b34d1c15f1aef62f6ea0a27",
	"0xa48a369d9e6ffbb0509e015203487d413aef500452357e441ac1d09079c27f086865d6a7546fb44215290b5029a580d6",
	"0x963a1b48ac60be0d27b207a2a60e5e8f3b43ec2160f15da865f7a5ed016eea94e8386d8f972571cdf277fe17890c6bdc",
	"0xad82a968b962b701d1e82c53da4c2959f7ee9ceb0f2ff0e06b03e66be9d12b32a9b3c04c090eb321369b3d7be159e2be",
	"0x89a04b671818e2a18df792177d6c3c72144a77f6857bef68c567bcf057c55239057255be21fcae27503e621ecb673664",
	"0x829e11d457851002fd08e48fdcf069286a16d2f5a00a11e3266953955b37ff40c242b72405d7ba5242ea3ef69d2aae33",
	"0x8bf72b083576e6dcea4196645c95a8ce60879fa2bfda4dfd15fc04be6cf48d9d42de9434b4f5f69b81d944c0228ec4f3",
}

// TODO: Not efficient, parsing all array in every call
func GetHardcodedConsensyscodefiKeys() [][]byte {
	depositedKeys := make([][]byte, 0)

	for _, keyStr := range consensuscodefiKeysStr {
		key, err := hexutil.Decode(keyStr)
		depositedKeys = append(depositedKeys, key)
		if err != nil {
			log.Fatal("Error parsing key: ", keyStr, ": ", err)
		}
	}
	return depositedKeys
}