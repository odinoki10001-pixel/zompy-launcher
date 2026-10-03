package mods

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Mod struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Source      string   `json:"source"` // modrinth | curseforge
	Downloads   int64    `json:"downloads"`
	Rating      float64  `json:"rating"`
	Updated     string   `json:"updated"`
	Categories  []string `json:"categories"`
	Versions    []string `json:"versions"`
	Loaders     []string `json:"loaders"`
	IconURL     string   `json:"iconUrl"`
	DownloadURL string   `json:"downloadUrl"`
}

var httpClient = &http.Client{Timeout: 20 * time.Second}

// ============ MODRINTH ============

const modrinthAPI = "https://api.modrinth.com/v2"

type mrHit struct {
	ProjectID    string   `json:"project_id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Author       string   `json:"author"`
	Downloads    int64    `json:"downloads"`
	IconURL      string   `json:"icon_url"`
	Categories   []string `json:"categories"`
	Versions     []string `json:"versions"`
	DateModified string   `json:"date_modified"`
}

type mrSearch struct {
	Hits []mrHit `json:"hits"`
}

func SearchModrinth(query, loader, gameVersion string) ([]Mod, error) {
	facets := [][]string{}
	if loader != "" && loader != "vanilla" {
		facets = append(facets, []string{"categories:" + loader})
	}
	if gameVersion != "" {
		facets = append(facets, []string{"versions:" + gameVersion})
	}
	q := url.Values{}
	q.Set("query", query)
	q.Set("limit", "30")
	if len(facets) > 0 {
		b, _ := json.Marshal(facets)
		q.Set("facets", string(b))
	}
	resp, err := httpClient.Get(modrinthAPI + "/search?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("modrinth: %s", resp.Status)
	}
	var out mrSearch
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	res := make([]Mod, 0, len(out.Hits))
	for _, h := range out.Hits {
		res = append(res, Mod{
			ID:          h.ProjectID,
			Name:        h.Title,
			Author:      h.Author,
			Description: h.Description,
			Source:      "modrinth",
			Downloads:   h.Downloads,
			Rating:      4.7,
			Updated:     h.DateModified,
			Categories:  h.Categories,
			Versions:    h.Versions,
			IconURL:     h.IconURL,
		})
	}
	return res, nil
}

// ============ CURSEFORGE ============

const cfAPI = "https://api.curseforge.com/v1"

type cfSearchResp struct {
	Data []struct {
		ID            int    `json:"id"`
		Name          string `json:"name"`
		Summary       string `json:"summary"`
		DownloadCount int64  `json:"downloadCount"`
		Logo          struct {
			URL string `json:"thumbnailUrl"`
		} `json:"logo"`
		Authors []struct {
			Name string `json:"name"`
		} `json:"authors"`
		LatestFiles []struct {
			GameVersions []string `json:"gameVersions"`
			DownloadURL  string   `json:"downloadUrl"`
		} `json:"latestFiles"`
	} `json:"data"`
}

func SearchCurseForge(query, loader, gameVersion, apiKey string) ([]Mod, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("CurseForge API key не задан")
	}
	q := url.Values{}
	q.Set("gameId", "432")
	q.Set("searchFilter", query)
	q.Set("pageSize", "30")
	q.Set("sortField", "2")
	q.Set("sortOrder", "desc")
	if gameVersion != "" {
		q.Set("gameVersion", gameVersion)
	}

	req, _ := http.NewRequest("GET", cfAPI+"/mods/search?"+q.Encode(), nil)
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("curseforge: %s", resp.Status)
	}
	var out cfSearchResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	res := make([]Mod, 0, len(out.Data))
	for _, d := range out.Data {
		author := ""
		if len(d.Authors) > 0 {
			author = d.Authors[0].Name
		}
		var versions []string
		var dlURL string
		if len(d.LatestFiles) > 0 {
			versions = d.LatestFiles[0].GameVersions
			dlURL = d.LatestFiles[0].DownloadURL
		}
		res = append(res, Mod{
			ID:          fmt.Sprint(d.ID),
			Name:        d.Name,
			Author:      author,
			Description: d.Summary,
			Source:      "curseforge",
			Downloads:   d.DownloadCount,
			Rating:      4.7,
			Versions:    versions,
			IconURL:     d.Logo.URL,
			DownloadURL: dlURL,
		})
	}
	return res, nil
}

// ============ DOWNLOAD ============

func Download(m Mod, destDir string) error {
	if m.DownloadURL == "" {
		return fmt.Errorf("нет ссылки на скачивание для %s", m.Name)
	}
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	name := strings.ReplaceAll(m.Name, " ", "_") + ".jar"
	dest := filepath.Join(destDir, name)

	resp, err := httpClient.Get(m.DownloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download: %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
