package main

import "fmt"

type macProject struct {
	ID, Name, Summary                           string
	Repository, Revision                        string
	SetupURL, SetupName, SetupSource, SetupNote string
	Requirements                                string
}

var macProjects = []macProject{
	{
		ID:           "gods-eye-view",
		Name:         "God's Eye View",
		Summary:      "Explore live public signals on a photorealistic 3D globe.",
		Repository:   "https://github.com/bilawalsidhu/gods-eye-view",
		Revision:     "844c25212c06cfa26ff6c0f3cb82bc9c325f4a85",
		SetupURL:     "https://pinokio.co/apps/github-com-bilawalsidhu-gods-eye-view",
		SetupName:    "Open guided Pinokio setup",
		SetupSource:  "Pinokio / external setup",
		SetupNote:    "Pinokio owns this mutable external installation flow; review its prompts and permissions before continuing.",
		Requirements: "Runs locally in a browser. The manual path requires Node.js 24.14+ below 25, or Node.js 26. Optional provider keys can enable paid services.",
	},
	{
		ID:           "flectar-mail",
		Name:         "Flectar Mail",
		Summary:      "A lightweight native client for email, calendars and contacts.",
		Repository:   "https://github.com/flectar/mail",
		Revision:     "d666d950138c1e60f70cd1a5ea655f40c5ecc8b9",
		SetupURL:     "https://github.com/flectar/mail/releases",
		SetupName:    "Open preview releases",
		SetupSource:  "GitHub Releases / external preview install",
		SetupNote:    "Choose a pre-release deliberately and verify its published SHA-256 checksum or provenance. macOS previews are ad-hoc signed and not notarized, so security prompts are expected.",
		Requirements: "In development only. macOS previews require macOS 14+; Gmail and Outlook sign-in may need your own OAuth registration until provider verification is complete. IMAP, JMAP and standards-based accounts remain available.",
	},
	{
		ID:           "whiteboard-animator",
		Name:         "Whiteboard Animator",
		Summary:      "Turns whiteboard-style images into hand-drawn reveal videos.",
		Repository:   "https://github.com/masihsultani/whiteboard-animator",
		Revision:     "e6e4dbcfc06e65b82490323a78bd9c277a9e2a0b",
		SetupURL:     "https://pypi.org/project/whiteboard-animator/",
		SetupName:    "Open PyPI package setup",
		SetupSource:  "PyPI / external package setup",
		SetupNote:    "PyPI publishes mutable package releases; review the package version and installation output before continuing.",
		Requirements: "Requires Python 3.10+ plus ffmpeg and ffprobe. The package includes an approximately 83 MB text-detection model; core rendering needs no API key, while region detection can use an optional Gemini key.",
	},
}

func macProjectByID(id string) (macProject, bool) {
	for _, project := range macProjects {
		if project.ID == id {
			return project, true
		}
	}
	return macProject{}, false
}

func macProjectRows() []macRow {
	rows := make([]macRow, 0, len(macProjects))
	for _, project := range macProjects {
		rows = append(rows, macRow{
			ID:      project.ID,
			Name:    project.Name,
			Summary: project.Summary,
			Source:  "Open-source local project / GitHub",
			Note:    project.Requirements + "\n\nRepository: " + project.Repository + "\nReviewed revision: " + project.Revision + "\nMagus opens a reviewed setup path only; it does not clone source, install npm dependencies, store API keys or claim ownership of the project.",
		})
	}
	return rows
}

func macProjectActionRows(id string) []macRow {
	project, ok := macProjectByID(id)
	if !ok {
		return nil
	}
	return []macRow{
		{ID: project.SetupURL, Name: project.SetupName, Summary: "Open the project's recommended package or guided setup path.", Source: project.SetupSource, Note: project.Requirements + " " + project.SetupNote},
		{ID: project.Repository + "/tree/" + project.Revision, Name: "Open reviewed source and manual setup", Summary: "Inspect the exact reviewed code and its terminal instructions.", Source: "GitHub / pinned revision " + project.Revision, Note: fmt.Sprintf("The repository currently documents: git clone, npm ci, npm run doctor, then npm run dev. %s", project.Requirements)},
	}
}

func macProjectSetupURLAllowed(url string) bool {
	for _, project := range macProjects {
		for _, row := range macProjectActionRows(project.ID) {
			if row.ID == url {
				return true
			}
		}
	}
	return false
}
