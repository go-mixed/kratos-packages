package command

type EndOfWire struct {
}

func BindCommands(rootCmd ICmder, subCmds Commands) *EndOfWire {
	rootCmd.AddCommand(subCmds...)
	return nil
}
