package record

// Release describes a Proxmox VE installer ISO that the recording pipeline
// knows how to provision.
//
// The Proxmox project does not publish a "latest" manifest; bump these by hand
// when a new stable PVE major or maintenance ISO ships. Both URL and SHA256
// must be updated together — checksum mismatches abort the pipeline rather
// than silently produce a different cassette set.
type Release struct {
	// Major is the marketing major version: "pve8", "pve9".
	Major string

	// CassetteDir is the relative dir under tests/recorder/testdata/ that
	// cassettes for this release are written to.
	CassetteDir string

	// ISOURL is the upstream installer ISO URL.
	ISOURL string

	// ISOFilename is the name the ISO will be stored as in the outer PVE
	// host's storage (the "filename" parameter of /storage/{name}/download-url).
	ISOFilename string

	// ISOSHA256 is the expected SHA-256 hex digest of the ISO. Proxmox's
	// download-url endpoint will refuse to write the file if the actual
	// digest disagrees.
	ISOSHA256 string
}

// Releases is the supported set, keyed by Major.
//
// PVE 7 is intentionally absent: proxmox-auto-install-assistant landed in PVE
// 8.2, PVE 7 hit EOL in July 2024, and the existing tests/mocks/pve7x/* gock
// fixtures cover the legacy path until the next major go-proxmox bump.
//
// To bump a release: download the new ISO, verify the upstream SHA256 from
// the Proxmox downloads page, replace URL+filename+digest below, then run
// `mage record:all` against your outer PVE.
var Releases = map[string]Release{
	"pve9": {
		Major:       "pve9",
		CassetteDir: "pve9",
		ISOURL:      "https://enterprise.proxmox.com/iso/proxmox-ve_9.1-1.iso",
		ISOFilename: "proxmox-ve_9.1-1.iso",
		ISOSHA256:   "6d8f5afc78c0c66812d7272cde7c8b98be7eb54401ceb045400db05eb5ae6d22",
	},
	"pve8": {
		Major:       "pve8",
		CassetteDir: "pve8",
		ISOURL:      "https://enterprise.proxmox.com/iso/proxmox-ve_8.4-1.iso",
		ISOFilename: "proxmox-ve_8.4-1.iso",
		ISOSHA256:   "d237d70ca48a9f6eb47f95fd4fd337722c3f69f8106393844d027d28c26523d8",
	},
}
