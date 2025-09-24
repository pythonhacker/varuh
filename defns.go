package varuh

type ActionFunc func(string) error
type ActionFunc2 func(string) (error, string)
type VoidFunc func() error
type VoidFunc2 func() (error, string)
type SettingFunc func(string)

const VERSION = 0.41
const APP = "varuh"

const AUTHOR_INFO = `
AUTHORS
    Copyright (C) 2025 @skeptichacker Anand B Pillai
`
