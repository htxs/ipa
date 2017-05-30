package ipa

import (
	"archive/zip"
	"fmt"
	"howett.net/plist"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	SearchIcon  = "(?i).*.app/*icon*"
	SearchPlist = "(?i).*.app/*info.plist"
)

const (
	SubDirectoryIcons  = "icons"
	SubDirectoryPlists = "plists"
	SubDirectoryAssets = "assets"
)

const (
	AssetFileTypeIcon      = 0
	AssetFileTypeInfoPlist = 1
)

type AssetFileType int

type AssetFile struct {
	Path          string
	AssetFileType AssetFileType
}

type InfoPlist struct {
	BundleName         string `plist:"CFBundleName"`
	DisplayName        string `plist:"CFBundleDisplayName"`
	BundleVersion      string `plist:"CFBundleVersion"`
	ShortBundleVersion string `plist:"CFBundleShortVersionString"`
	BundleIdentifier   string `plist:"CFBundleIdentifier"`
}

type destinationPathStruct struct {
	Path string
}

type fileModelClass struct {
	Path string
	File *zip.File
	Size int64
}

func newFileModel(destinationParentPath string, index int, file *zip.File) fileModelClass {

	_, path := filepath.Split(file.Name)
	fileParts := strings.Split(path, ".")
	completeFileName := fileParts[0]
	if index != 0 {
		completeFileName += fmt.Sprintf("%v", index)
	}
	completeFileName += "." + fileParts[1]
	path = filepath.Join(destinationParentPath, completeFileName)

	return fileModelClass{Path: path, File: file}
}

func (f fileModelClass) String() string {

	parts := strings.Split(f.Path, "/")

	return fmt.Sprintf("%v", parts[len(parts)-1])
}

// sort by file size
type bySizeStruct []fileModelClass

func (b bySizeStruct) Len() int {
	return len(b)
}

func (b bySizeStruct) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b bySizeStruct) Less(i, j int) bool {
	return b[i].Size > b[j].Size
}

func createDestinationPathsIfNeccessary(destinationPaths map[string]destinationPathStruct) error {

	var err error
	for _, destinationPath := range destinationPaths {
		err = os.MkdirAll(destinationPath.Path, 0755)
	}

	return err
}

func deleteDestinationPaths(destinationPaths map[string]destinationPathStruct) error {

	var err error
	for _, destinationPath := range destinationPaths {
		err = os.RemoveAll(destinationPath.Path)
	}

	return err
}

func copyFileToFolder(destinationPath, sourcePath string) (destinationFilePath string, err error) {

	filePath := strings.Split(sourcePath, "/")
	fileName := filePath[len(filePath)-1]

	destinationFilePath = filepath.Join(destinationPath, fileName)
	err = copyFileToFile(destinationFilePath, sourcePath)

	return
}

func copyFileToFile(destinationPath, sourcePath string) error {

	sf, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer df.Close()

	if _, err := io.Copy(df, sf); err != nil {
		return err
	}

	return nil
}

func unzip(ipaPath string, destinationPaths map[string]destinationPathStruct) error {

	r, err := zip.OpenReader(ipaPath)
	if err != nil {
		return err
	}

	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	extractAndWriteFile := func(fileModel fileModelClass) error {
		file := fileModel.File
		path := fileModel.Path

		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err = rc.Close(); err != nil {
				panic(err)
			}
		}()

		if file.FileInfo().IsDir() {
			os.Mkdir(path, file.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	var extractionFiles []fileModelClass
	plistIndex := 0
	iconIndex := 0
	for _, f := range r.File {
		if matched, _ := regexp.MatchString(SearchPlist, f.Name); matched {
			extractionFiles = append(extractionFiles, newFileModel(destinationPaths[SubDirectoryPlists].Path, plistIndex, f))
			plistIndex++
		} else if matched, _ := regexp.MatchString(SearchIcon, f.Name); matched {
			extractionFiles = append(extractionFiles, newFileModel(destinationPaths[SubDirectoryIcons].Path, iconIndex, f))
			iconIndex++
		}
	}

	for _, fileModel := range extractionFiles {
		err := extractAndWriteFile(fileModel)
		if err != nil {
			return err
		}
	}

	return nil
}

func ExtractInformationForIpaWithPath(ipaPath string, destinationPath string) (error error, infoPlist *InfoPlist, assets []AssetFile) {

	plistPathStruct := destinationPathStruct{Path: filepath.Join(destinationPath, "/"+SubDirectoryPlists+"/")}
	iconPathStruct := destinationPathStruct{Path: filepath.Join(destinationPath, "/"+SubDirectoryIcons+"/")}
	assetsPathStruct := destinationPathStruct{Path: filepath.Join(destinationPath, "/"+SubDirectoryAssets+"/")}
	destinationPaths := map[string]destinationPathStruct{SubDirectoryPlists: plistPathStruct, SubDirectoryIcons: iconPathStruct, SubDirectoryAssets: assetsPathStruct}

	var assetFiles []AssetFile

	err := createDestinationPathsIfNeccessary(destinationPaths)
	if err != nil {
		fmt.Println(err)
	}
	err = unzip(ipaPath, destinationPaths)

	plistFolder := destinationPaths[SubDirectoryPlists].Path
	plistFolderContents, err := ioutil.ReadDir(plistFolder)
	if err != nil {
		fmt.Println(err)
	}

	var info InfoPlist
	for _, plistFile := range plistFolderContents {
		filePath := filepath.Join(plistFolder, plistFile.Name())
		p, err := os.Open(filePath)
		if err != nil {
			fmt.Println(err)
		}
		defer p.Close()

		decoder := plist.NewDecoder(p)
		err = decoder.Decode(&info)
		if err != nil {
			fmt.Println(err)
		}
		if len(info.BundleIdentifier) != 0 {
			destinationFilePath, err := copyFileToFolder(filepath.Join(destinationPath, SubDirectoryAssets), filePath)
			assetFiles = append(assetFiles, AssetFile{Path: destinationFilePath, AssetFileType: AssetFileTypeInfoPlist})
			if err != nil {
				fmt.Println(err)
			}
			break
		}
	}

	iconFolder := destinationPaths[SubDirectoryIcons].Path
	iconFolderContents, err := ioutil.ReadDir(iconFolder)
	if err != nil {
		fmt.Println(err)
	}

	var icons []fileModelClass
	for _, iconFile := range iconFolderContents {
		icons = append(icons, fileModelClass{Path: filepath.Join(iconFolder, iconFile.Name()), Size: iconFile.Size()})
	}

	sort.Sort(bySizeStruct(icons))
	if len(icons) > 0 {
		destinationFilePath, err := copyFileToFolder(filepath.Join(destinationPath, SubDirectoryAssets), icons[0].Path)
		assetFiles = append(assetFiles, AssetFile{Path: destinationFilePath, AssetFileType: AssetFileTypeIcon})
		if err != nil {
			fmt.Println(err)
		}
	}

	// delete temporary files except assets
	delete(destinationPaths, SubDirectoryAssets)
	err = deleteDestinationPaths(destinationPaths)
	if err != nil {
		fmt.Println(err)
	}

	if err != nil {
		return err, nil, nil
	} else {
		return nil, &info, assetFiles
	}
}
