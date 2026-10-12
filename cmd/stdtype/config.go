package main

import "github.com/spf13/pflag"

func SetSettingsFlag(flags *pflag.FlagSet) {
	flags.String("mode", "", "test mode: time or words")
	flags.Int("duration", 0, "time-mode duration in seconds")
	flags.Int("words", 0, "words-mode word count")
	flags.String("list", "", "word list (e.g. english, english_1k)")
	flags.Int("lines", 0, "visible passage lines")
	flags.Bool("punctuation", false, "enable punctuation")
	flags.Bool("numbers", false, "enable numbers")
}
