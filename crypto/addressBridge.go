package crypto

import "strings"

//"drep3726430e6E3448753F29E1d3ca24120850433Ff9"

func DrepToEth(drepAddress string) (ethAddress CommonAddress) {

	drepAddress = strings.ToLower(drepAddress)
	if strings.HasPrefix(drepAddress, "drep") {
		drepAddress = "0x" + drepAddress[4:]
	}

	ethAddress = HexToAddress(drepAddress)
	return
}

func EthToDrep(ethAddress *CommonAddress) (drepAddress string) {
	return "DREP" + ethAddress.String()[2:]
}
