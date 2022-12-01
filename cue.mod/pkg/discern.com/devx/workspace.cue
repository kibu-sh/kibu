package devx

#ConfigKey: {
	Env: "development" | "staging" | "production"
	Engine: "hashivault" | "gcpkms" | "awskms" | "azurevaultkey"
	Path: string
}

#ConfigStoreSettings: {
	Keys: [#ConfigKey]
}

#Workspace: {
	ConfigStore?: #ConfigStoreSettings
}