# TacticalRMM Linux Agent Manager

Or trmm-lam for short.

## _What does this do?_

It manages your Tactical RMM agent on Linux systems. It can perform new installs in addition to updating existing installs.

For installs, it essentially just logs into your Tactical RMM server (and mesh central server), generates the installer script, and runs it. The installer script will install the agent on your system and configure it to communicate with your Tactical RMM server like usual.

For updates, it downloads the latest binary from my [GitHub releases page](https://github.com/SoarinFerret/rmmagent-builder) and replaces the existing binary with the new one. You can swap out the url with a commandline flag if you would like to use a different binary. All my binaries are built from the official source code (you can read my github actions), so you should be able to trust them.

## _Why did you make this?_

I wanted to install the TacticalRMM agent on my non-nixos Linux systems, and I wanted it to be as easy as possible. I also wanted to be able to easily update the agent when new versions are released.

## _How do I use this?_

You can download the latest release from the [releases page](https://github.com/soarinferret/trmm-lam/releases).

You can also build it yourself by cloning the repository and running `CGO_ENABLED=0 go build -ldflags="-s -w"`.

## _Does this require sponsorship or a license to use?_

No, this is using unofficial binaries and is not affiliated with Tactical RMM. You can use this without a license or sponsorship - but please consider sponsoring Amidaware / Tactical RMM if you are able to.

## _Is this installer official or affiliated with Amidaware LLC / Tactical RMM?_

Nope. This is a personal project that I made for my own use. I am not affiliated with Amidaware LLC or Tactical RMM in any way (besides being a paying customer through my employer).

## _Why is this written in golang instead of X scripting language?_

Because I can statically compile the binary and distribute it without worrying about dependencies. Also, I like writing small utilities in golang for fun.

## License

MIT licensed - see [LICENSE](LICENSE) for more information.
