package main

type macBrowserAddonLink struct{ Name, URL string }

type macBrowserAddon struct {
	ID, Name, Summary, Note string
	Links                   []macBrowserAddonLink
}

var macBrowserAddons = []macBrowserAddon{
	{
		ID:      "your-dynamic-dashboard",
		Name:    "YourDynamicDashboard",
		Summary: "A local-first, customizable dashboard for new tabs.",
		Note:    "Settings, tasks and search history stay in the browser. Optional weather, news and online suggestions contact their providers only when enabled.",
		Links: []macBrowserAddonLink{
			{"Chrome Web Store", "https://chromewebstore.google.com/detail/fckmlnagohleefboaleepppikpdkckjn"},
			{"Firefox Add-ons", "https://addons.mozilla.org/en-US/firefox/addon/yourdynamicdashboard/"},
			{"Microsoft Edge Add-ons", "https://microsoftedge.microsoft.com/addons/detail/yourdynamicdashboard/phhofebhbmicnfhmmdgikiddaboljnec"},
		},
	},
	{
		ID:      "ublock-origin",
		Name:    "uBlock Origin",
		Summary: "Efficient content blocking for Firefox.",
		Note:    "Review filter lists and site breakage before relying on it. Chromium browsers use different supported builds.",
		Links:   []macBrowserAddonLink{{"Firefox Add-ons", "https://addons.mozilla.org/firefox/addon/ublock-origin/"}},
	},
	{
		ID:      "dark-reader",
		Name:    "Dark Reader",
		Summary: "Adds adjustable dark themes to websites.",
		Note:    "It needs permission to change page appearance. Exclude sites that do not render correctly.",
		Links:   []macBrowserAddonLink{{"Firefox Add-ons", "https://addons.mozilla.org/firefox/addon/darkreader/"}},
	},
	{
		ID:      "bitwarden",
		Name:    "Bitwarden",
		Summary: "Open-source password manager with browser autofill.",
		Note:    "Sign in or create an account inside Bitwarden; configure its browser integration after installation.",
		Links:   []macBrowserAddonLink{{"Firefox Add-ons", "https://addons.mozilla.org/firefox/addon/bitwarden-password-manager/"}},
	},
	{
		ID:      "multi-account-containers",
		Name:    "Firefox Multi-Account Containers",
		Summary: "Keeps selected sites in separate browser identities.",
		Note:    "Create containers and assign sites after installation. It is specific to Firefox.",
		Links:   []macBrowserAddonLink{{"Firefox Add-ons", "https://addons.mozilla.org/firefox/addon/multi-account-containers/"}},
	},
	{
		ID:      "sponsorblock",
		Name:    "SponsorBlock",
		Summary: "Skips community-marked sponsored segments in videos.",
		Note:    "Segment data comes from the community; review its privacy and contribution settings after installation.",
		Links:   []macBrowserAddonLink{{"Firefox Add-ons", "https://addons.mozilla.org/firefox/addon/sponsorblock/"}},
	},
}

func macBrowserAddonByID(id string) (macBrowserAddon, bool) {
	for _, addon := range macBrowserAddons {
		if addon.ID == id {
			return addon, true
		}
	}
	return macBrowserAddon{}, false
}

func macBrowserAddonRows() []macRow {
	rows := make([]macRow, 0, len(macBrowserAddons))
	for _, addon := range macBrowserAddons {
		rows = append(rows, macRow{ID: addon.ID, Name: addon.Name, Summary: addon.Summary, Source: "Browser add-on / reviewed store links", Note: addon.Note})
	}
	return rows
}

func macBrowserAddonLinkRows(id string) []macRow {
	addon, ok := macBrowserAddonByID(id)
	if !ok {
		return nil
	}
	rows := make([]macRow, 0, len(addon.Links))
	for _, link := range addon.Links {
		rows = append(rows, macRow{ID: link.URL, Name: "Open " + link.Name, Summary: "Open the official listing for " + addon.Name + ".", Source: link.Name, Note: addon.Note + " Browser stores manage installation and updates; review the requested permissions before adding it."})
	}
	return rows
}

func macBrowserAddonURLAllowed(url string) bool {
	for _, addon := range macBrowserAddons {
		for _, row := range macBrowserAddonLinkRows(addon.ID) {
			if row.ID == url {
				return true
			}
		}
	}
	return false
}
