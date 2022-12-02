package devx

#EncryptionKey: {
	Engine: "hashivault" | "gcpkms" | "awskms" | "azurevaultkey"
	Env: string
	Key: string
}

#ConfigStoreSettings: {
	EncryptionKeys: [...#EncryptionKey]
}

#Workspace: {
	ConfigStore?: #ConfigStoreSettings
}