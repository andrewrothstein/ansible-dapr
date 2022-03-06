package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Ver struct {
	Major int
	Minor int
	Patch int
}

func (v *Ver) Fmt() string {
	return fmt.Sprintf("%v.%v.%v", v.Major, v.Minor, v.Patch)
}

type Platform struct {
	Os          string
	Arch        string
	ArchiveType string
}

func NewPlatformTGZ(os string, arch string) Platform {
	return Platform{Os: os, Arch: arch, ArchiveType: "tar.gz"}
}

func NewPlatformZIP(os string, arch string) Platform {
	return Platform{Os: os, Arch: arch, ArchiveType: "zip"}
}

func (s *Platform) Fmt() string {
	return fmt.Sprintf("%s%s%s", s.Os, "_", s.Arch)
}

type Params struct {
	Mirror string
}

func dl_url(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New("not found")
	}
	defer resp.Body.Close()
	if b, err := io.ReadAll(resp.Body); err == nil {
		return string(b), nil
	}
	return "", err
}

func dl_checksum(checksum_url string, f string) (string, error) {
	s, err := dl_url(checksum_url)
	if err == nil {
		lines := strings.Split(s, "\n")
		for _, line := range lines {
			sums := strings.Fields(line)
			if len(sums) > 1 && strings.HasSuffix(sums[1], f) {
				return sums[0], nil
			}
		}
	}
	return "", err
}

// https://github.com/dapr/cli/releases/download/v1.6.0/dapr_linux_amd64.tar.gz.sha256

func dl(
	params *Params,
	app string,
	vs []Ver,
	platforms []Platform,
) {
	for _, v := range vs {
		fmt.Printf("  '%s':\n", v.Fmt())
		for _, p := range platforms {
			file := fmt.Sprintf(
				"%s_%s.%s",
				app, p.Fmt(), p.ArchiveType,
			)
			checksumsurl := fmt.Sprintf(
				"%s/v%s/%s.sha256",
				params.Mirror, v.Fmt(), file,
			)
			if checksum, err := dl_checksum(checksumsurl, file); err == nil {
				fmt.Printf("    # %s\n", checksumsurl)
				fmt.Printf("    %s: sha256:%s\n", p.Fmt(), checksum)
			}
		}
	}
}

func main() {
	params := Params{
		Mirror: "https://github.com/dapr/cli/releases/download",
	}

	platforms := []Platform{
		NewPlatformTGZ("darwin", "amd64"),
		NewPlatformTGZ("darwin", "arm64"),
		NewPlatformTGZ("linux", "amd64"),
		NewPlatformTGZ("linux", "arm"),
		NewPlatformTGZ("linux", "arm64"),
		NewPlatformZIP("windows", "amd64"),
	}

	versions := []Ver{
		{Major: 1, Minor: 6, Patch: 0},
	}
	dl(&params, "dapr", versions, platforms)
}
