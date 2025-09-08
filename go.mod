module github.com/Sumire-Labs/Nyx

go 1.24.1

toolchain go1.24.4

require (
	github.com/Sumire-Labs/Nyx-API v0.3.1
	github.com/bwmarrin/discordgo v0.29.0
	github.com/mattn/go-sqlite3 v1.14.17
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)

replace github.com/Sumire-Labs/Nyx-API => ../Nyx-API
