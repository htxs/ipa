package ipa

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExtractInformationForIpaWithPath(t *testing.T) {
	
	ipaPath := "../testfile/BrainCoWord.ipa"
	destinationPath := "../testfile"
	
	error, infoPlist, assets := ExtractInformationForIpaWithPath(ipaPath, destinationPath)
	
	assert.Nil(t, error)
	
	assert.NotNil(t, infoPlist)
	assert.Equal(t, "BrainCoWord", infoPlist.BundleName)
	assert.Equal(t, "BrainCo课堂", infoPlist.DisplayName)
	assert.Equal(t, "8", infoPlist.BundleVersion)
	assert.Equal(t, "1.0", infoPlist.ShortBundleVersion)
	assert.Equal(t, "cn.com.andconsulting.BrainCoWord", infoPlist.BundleIdentifier)
	
	assert.NotEmpty(t, assets)
	assert.Len(t, assets, 2)
	
	asset0 := assets[0]
	assert.NotNil(t, asset0)
	assert.NotEmpty(t, asset0.Path)
	assert.Subset(t, []int{0, 1}, []int{int(asset0.AssetFileType)})
	
	asset1 := assets[1]
	assert.NotNil(t, asset1)
	assert.NotEmpty(t, asset1.Path)
	assert.Subset(t, []int{0, 1}, []int{int(asset1.AssetFileType)})
}