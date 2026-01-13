# Get-the-mail

<img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/go/go-original.svg" alt="Go icon" width="32" height="32" /> <img src="https://cdn.jsdelivr.net/gh/devicons/devicon@latest/icons/yaml/yaml-original.svg" alt="YAML icon" width="32" height="32" />

A simple Go app to collect email domain from various sources and output them in a digestible format.

**Why?** I wanted a simple way to gather potentially unwanted email domains, but all the other tools seem overengineered.

## Features
- Collect email domains from multiple sources (currently supports only txt files)
- Whitelist functionality to exclude certain domains
- Output results in txt files
- Easy configuration via YAML file
- Lightweight and fast
- Comes with predefined sources

## Running
```shell
go run main.go
```
