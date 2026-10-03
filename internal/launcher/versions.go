package versions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const manifestURL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

type Version struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // release | snapshot | old_beta | old_alpha
	URL       string `json:"url"`
	Time      string `json:"releaseTime"`
	Installed bool   `json:"installed"`
}

type manifest struct {
	Latest   map[string]string `json:"latest"`
	Versions []Version         `json:"versions"`
}

func manifestPath(dataDir string) string {
	return filepath.Join(dataDir, "versions_manifest.json")
}

func EnsureManifest(dataDir string) (string, error) {
	p := manifestPath(dataDir)
	if st, err := os.Stat(p); err == nil {
		if time.Since(st.ModTime()) < 6*time.Hour {
			return p, nil
		}
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return "", err
	}
	resp, err := http.Get(manifestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	out, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return p, err
}

func List(dataDir string) ([]Version, error) {
	p, err := EnsureManifest(dataDir)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	// помечаем установленные
	for i := range m.Versions {
		if _, err := os.Stat(filepath.Join(dataDir, "versions", m.Versions[i].ID)); err == nil {
			m.Versions[i].Installed = true
		}
	}
	return m.Versions, nil
}

func Install(dataDir, id string) error {
	list, err := List(dataDir)
	if err != nil {
		return err
	}
	var target *Version
	for i := range list {
		if list[i].ID == id {
			target = &list[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("version %s not found", id)
	}

	dest := filepath.Join(dataDir, "versions", id)
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	resp, err := http.Get(target.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath.Join(dest, id+".json"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
