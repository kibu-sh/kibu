package devx

#ConfigKey: {
	Env: "development" | "staging" | "production"
	Engine: "hashivault" | "gcpkms" | "awskms" | "azurevaultkey"
	Key: string
}

#ConfigStoreSettings: {
	Keys: [#ConfigKey]
}

#Workspace: {
	ConfigStore?: #ConfigStoreSettings
}