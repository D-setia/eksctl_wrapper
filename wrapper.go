// wrapper.go
package main

import (
    "C"
    "bytes"
    "fmt"
    "io"
	"os"
	"strings"
	"time"
    // "os"

	"github.com/kris-nova/logger"
    "github.com/spf13/cobra"
    "github.com/fatih/color"
    "github.com/spf13/pflag"
    lol "github.com/kris-nova/lolgopher"

	// "github.com/weaveworks/eksctl/pkg/ctl/completion"
    "github.com/weaveworks/eksctl/pkg/ctl/create"
    "github.com/weaveworks/eksctl/pkg/ctl/cmdutils"
	"github.com/weaveworks/eksctl/pkg/info"
    "github.com/weaveworks/eksctl/pkg/version"


)

func infoCmd(cmd *cmdutils.Cmd) {
	var output string

	cmd.SetDescription("info", "Output the version of eksctl, kubectl and OS info", "")
	cmd.FlagSetGroup.InFlagSet("General", func(fs *pflag.FlagSet) {
		fs.StringVarP(&output, "output", "o", "", "specifies the output format (valid option: json)")
	})
	cmd.CobraCommand.Args = cobra.NoArgs
	cmd.CobraCommand.RunE = func(_ *cobra.Command, args []string) error {
		switch output {
		case "":
			version := info.GetInfo()
			fmt.Printf("eksctl version: %s\n", version.EksctlVersion)
			fmt.Printf("kubectl version: %s\n", version.KubectlVersion)
			fmt.Printf("OS: %s\n", version.OS)
		case "json":
			fmt.Printf("%s\n", info.String())
		default:
			return fmt.Errorf("unknown output: %s", output)
		}
		return nil
	}
}

func initLogger(level int, colorValue string, logBuffer *bytes.Buffer, dumpLogsValue bool) {
	logger.Layout = "2006-01-02 15:04:05"

	var bitwiseLevel int
	switch level {
	case 4:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical | logger.LogWarning | logger.LogInfo | logger.LogDebug
	case 3:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical | logger.LogWarning | logger.LogInfo
	case 2:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical | logger.LogWarning
	case 1:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical
	case 0:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess
	default:
		bitwiseLevel = logger.LogDeprecated | logger.LogEverything
	}
	logger.BitwiseLevel = bitwiseLevel

	if dumpLogsValue {
		switch colorValue {
		case "fabulous":
			logger.Writer = io.MultiWriter(lol.NewLolWriter(), logBuffer)
		case "true":
			logger.Writer = io.MultiWriter(color.Output, logBuffer)
		default:
			logger.Writer = io.MultiWriter(os.Stdout, logBuffer)
		}

	} else {
		switch colorValue {
		case "fabulous":
			logger.Writer = lol.NewLolWriter()
		case "true":
			logger.Writer = color.Output
		default:
			logger.Writer = os.Stdout
		}
	}

	logger.Line = func(prefix, format string, a ...interface{}) string {
		if !strings.Contains(format, "\n") {
			format = fmt.Sprintf("%s%s", format, "\n")
		}
		now := time.Now()
		fNow := now.Format(logger.Layout)
		var colorize func(format string, a ...interface{}) string
		var icon string
		switch prefix {
		case logger.PreAlways:
			icon = "✿"
			colorize = color.GreenString
		case logger.PreCritical:
			icon = "✖"
			colorize = color.RedString
		case logger.PreInfo:
			icon = "ℹ"
			colorize = color.CyanString
		case logger.PreDebug:
			icon = "▶"
			colorize = color.GreenString
		case logger.PreSuccess:
			icon = "✔"
			colorize = color.CyanString
		case logger.PreWarning:
			icon = "!"
			colorize = color.GreenString
		default:
			icon = "ℹ"
			colorize = color.CyanString
		}

		out := fmt.Sprintf(format, a...)
		out = fmt.Sprintf("%s [%s]  %s", fNow, icon, out)
		if colorValue == "true" {
			out = colorize(out)
		}

		return out
	}
}


func versionCmd(cmd *cmdutils.Cmd) {
	var output string

	cmd.SetDescription("version", "Output the version of eksctl", "")
	cmd.FlagSetGroup.InFlagSet("General", func(fs *pflag.FlagSet) {
		fs.StringVarP(&output, "output", "o", "", "specifies the output format (valid option: json)")
	})
	cmd.CobraCommand.Args = cobra.NoArgs
	cmd.CobraCommand.RunE = func(_ *cobra.Command, args []string) error {
		switch output {
		case "":
			fmt.Printf("%s\n", version.GetVersion())
		case "json":
			fmt.Printf("%s\n", version.String())
		default:
			return fmt.Errorf("unknown output: %s", output)
		}
		return nil
	}
}

func addCommands(rootCmd *cobra.Command, flagGrouping *cmdutils.FlagGrouping) {
	rootCmd.AddCommand(create.Command(flagGrouping))
	//Ensures "eksctl --help" presents eksctl anywhere as a command, but adds no subcommands since we invoke the binary.
	rootCmd.AddCommand(cmdutils.NewVerbCmd("anywhere", "EKS anywhere", ""))

	cmdutils.AddResourceCmd(flagGrouping, rootCmd, infoCmd)
	cmdutils.AddResourceCmd(flagGrouping, rootCmd, versionCmd)
}

func checkCommand(rootCmd *cobra.Command) {
	for _, cmd := range rootCmd.Commands() {
		// just a precaution as the verb command didn't have runE
		if cmd.RunE != nil {
			continue
		}
		cmd.RunE = func(c *cobra.Command, args []string) error {
			var e error
			if len(args) == 0 {
				e = fmt.Errorf("please provide a valid resource for \"%s\"", c.Name())
			} else {
				e = fmt.Errorf("unknown resource type \"%s\"", args[0])
			}
			fmt.Printf("Error: %s\n\n", e.Error())

			if err := c.Help(); err != nil {
				logger.Debug("ignoring cobra error %q", err.Error())
			}
			return e
		}
	}
}


//export CreateCluster
func CreateCluster(configFile *C.char) *C.char {
    rootCmd := &cobra.Command{
        Use:   "eksctl [command]",
        Short: "The official CLI for Amazon EKS",
        Run: func(c *cobra.Command, _ []string) {
            if err := c.Help(); err != nil {
                fmt.Printf("ignoring cobra error %q\n", err.Error())
            }
        },
        SilenceUsage: true,
    }

    flagGrouping := cmdutils.NewGrouping()
    addCommands(rootCmd, flagGrouping)
    checkCommand(rootCmd)

    rootCmd.SetArgs([]string{"create", "cluster", "-f", C.GoString(configFile)})

    logBuffer := new(bytes.Buffer)
    cobra.OnInitialize(func() {
        initLogger(3, "true", logBuffer, false)
    })

    if err := rootCmd.Execute(); err != nil {
        return C.CString(err.Error())
    }

    return C.CString("Cluster created successfully")
}

func main() {}
