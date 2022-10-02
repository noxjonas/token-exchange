package logs

import (
	"flag"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

func AddFlags(fs *pflag.FlagSet) {
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		pf := pflag.PFlagFromGoFlag(f)
		if fs.Lookup(pf.Name) == nil {
			switch pf.Name {
			case "v":
				fs.AddFlag(pf)
			}
		}
	})
}

func init() {
	var allFlags flag.FlagSet
	klog.InitFlags(&allFlags)

	allFlags.VisitAll(func(f *flag.Flag) {
		switch f.Name {
		case "v":
			flag.CommandLine.Var(f.Value, f.Name, f.Usage)
		}
	})
}
