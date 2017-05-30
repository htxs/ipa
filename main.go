package main

import (
	"fmt"
	"htxs.me/ipa/ipa"
)

func main() {
	ipaPath := "./testfile/BrainCoWord.ipa"
	destinationPath := "./testfile"
	
	_, infoPlist, assets := ipa.ExtractInformationForIpaWithPath(ipaPath, destinationPath)
	fmt.Println(infoPlist)
	fmt.Println(assets)
}
