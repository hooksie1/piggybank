# gupdate

Create self updating binaries

## Background

This package wraps Minio's self updater package and allows for automatic downloading of Go binaries.

## Usage

Create a project and then use gupdate to get the latest release and update the binary.

```
gh := gupdate.GitHubProject{
	Name:     "coverwhale-go",
	Owner:    "CoverWhale",
	Platform: runtime.GOOS,
	Arch:     runtime.GOARCH,
	ChecksumFunc: gupdate.GoReleaserChecksum,
}

release, err := gupdate.GetLatestRelease(gh)
if err != nil {
	log.Fatal(err)
}

if err := release.Update(); err != nil {
	log.Fatal(err)
}
```

## Private Repos

For GitHub projects there is a Token field on the `GitHubProject` struct. This will automatically add the token to the request and use the correct headers.

It sets a `RequestFunc` under the hood which sets the token and accept types. This can be overridden if needed on both the GitHubProject and the Release.
