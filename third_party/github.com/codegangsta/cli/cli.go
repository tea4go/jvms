package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ErrWriter 用于输出错误信息，默认指向 os.Stderr。
var ErrWriter io.Writer = os.Stderr

// OutWriter 用于输出普通信息，默认指向 os.Stdout。
var OutWriter io.Writer = os.Stdout

// App 定义一个 CLI 应用。
type App struct {
	Name            string
	Usage           string
	Version         string
	Metadata        map[string]interface{}
	Commands        []Command
	Action          func(*Context) error
	CommandNotFound func(*Context, string)
}

// Command 表示一个具体的子命令。
type Command struct {
	Name    string
	Aliases []string
	Usage   string
	Flags   []Flag
	Action  func(*Context) error
}

// Flag 是命令行标志的统一接口。
type Flag interface{}

// StringFlag 定义字符串类型标志。
type StringFlag struct {
	Name  string
	Usage string
	Value string
}

// BoolFlag 定义布尔类型标志。
type BoolFlag struct {
	Name  string
	Usage string
	Value bool
}

// Context 保存命令执行时的上下文信息。
type Context struct {
	App     *App
	Command *Command
	args    []string
	flags   *flagSet
}

// Args 表示命令行中的位置参数集合。
type Args struct {
	args []string
}

// NewApp 创建一个 CLI 应用的默认实例。
func NewApp() *App {
	return &App{
		Metadata: map[string]interface{}{},
	}
}

// Run 执行 CLI 应用。
func (a *App) Run(arguments []string) error {
	if a.Metadata == nil {
		a.Metadata = map[string]interface{}{}
	}
	if len(arguments) == 0 {
		return errors.New("缺少执行参数")
	}

	var (
		globalArgs  []string
		commandName string
		commandArgs []string
	)

	for i := 1; i < len(arguments); i++ {
		arg := arguments[i]
		if commandName == "" {
			if arg == "--" {
				if i+1 < len(arguments) {
					commandArgs = append(commandArgs, arguments[i+1:]...)
				}
				break
			}
			if strings.HasPrefix(arg, "-") {
				globalArgs = append(globalArgs, arg)
				continue
			}
			commandName = arg
			if i+1 < len(arguments) {
				commandArgs = append(commandArgs, arguments[i+1:]...)
			}
			break
		}
	}

	if commandName == "" {
		for _, arg := range globalArgs {
			switch arg {
			case "--help", "-h":
				if a.Action != nil {
					return a.Action(newContext(a, nil, []string{}, newFlagSet(nil)))
				}
				return nil
			case "--version", "-v":
				fmt.Fprintln(OutWriter, a.Version)
				return nil
			default:
				// 未知全局选项直接忽略。
			}
		}
		if a.Action != nil {
			return a.Action(newContext(a, nil, []string{}, newFlagSet(nil)))
		}
		return nil
	}

	cmd := a.findCommand(commandName)
	if cmd == nil {
		ctx := newContext(a, nil, []string{}, newFlagSet(nil))
		if a.CommandNotFound != nil {
			a.CommandNotFound(ctx, commandName)
			return nil
		}
		return fmt.Errorf("未知命令: %s", commandName)
	}

	fs := newFlagSet(cmd.Flags)
	remaining, err := fs.parse(commandArgs)
	if err != nil {
		return err
	}

	ctx := newContext(a, cmd, remaining, fs)
	if cmd.Action != nil {
		return cmd.Action(ctx)
	}
	return nil
}

// Args 返回位置参数集合。
func (c *Context) Args() Args {
	return Args{args: append([]string{}, c.args...)}
}

// NArg 返回位置参数的数量。
func (c *Context) NArg() int {
	return len(c.args)
}

// String 返回字符串类型标志的值。
func (c *Context) String(name string) string {
	return c.flags.getString(name)
}

// Bool 返回布尔类型标志的值。
func (c *Context) Bool(name string) bool {
	return c.flags.getBool(name)
}

// IsSet 判断某个标志是否被显式设置过。
func (c *Context) IsSet(name string) bool {
	return c.flags.isSet(name)
}

// Present 判断是否存在位置参数。
func (a Args) Present() bool {
	return len(a.args) > 0
}

// First 返回第一个位置参数。
func (a Args) First() string {
	if len(a.args) == 0 {
		return ""
	}
	return a.args[0]
}

// Get 返回指定索引的参数。
func (a Args) Get(index int) string {
	if index < 0 || index >= len(a.args) {
		return ""
	}
	return a.args[index]
}

func (a *App) findCommand(name string) *Command {
	lower := strings.ToLower(name)
	for i := range a.Commands {
		cmd := &a.Commands[i]
		if strings.ToLower(cmd.Name) == lower {
			return cmd
		}
		for _, alias := range cmd.Aliases {
			if strings.ToLower(alias) == lower {
				return cmd
			}
		}
	}
	return nil
}

type flagSet struct {
	stringValues map[string]string
	boolValues   map[string]bool
	setFlags     map[string]bool
	alias        map[string]string
}

func newFlagSet(flags []Flag) *flagSet {
	fs := &flagSet{
		stringValues: map[string]string{},
		boolValues:   map[string]bool{},
		setFlags:     map[string]bool{},
		alias:        map[string]string{},
	}

	for _, fl := range flags {
		switch v := fl.(type) {
		case StringFlag:
			fs.registerStringFlag(&v)
		case *StringFlag:
			fs.registerStringFlag(v)
		case BoolFlag:
			fs.registerBoolFlag(&v)
		case *BoolFlag:
			fs.registerBoolFlag(v)
		default:
			// 未知类型忽略
		}
	}

	return fs
}

func (fs *flagSet) registerStringFlag(flag *StringFlag) {
	names := parseFlagNames(flag.Name)
	if len(names) == 0 {
		return
	}
	canonical := names[0]
	fs.stringValues[canonical] = flag.Value
	for _, n := range names {
		fs.alias[n] = canonical
	}
}

func (fs *flagSet) registerBoolFlag(flag *BoolFlag) {
	names := parseFlagNames(flag.Name)
	if len(names) == 0 {
		return
	}
	canonical := names[0]
	fs.boolValues[canonical] = flag.Value
	for _, n := range names {
		fs.alias[n] = canonical
	}
}

func (fs *flagSet) parse(args []string) ([]string, error) {
	if fs == nil {
		return args, nil
	}
	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); {
		arg := args[i]
		if arg == "--" {
			remaining = append(remaining, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			remaining = append(remaining, arg)
			i++
			continue
		}

		namePart, valuePart, hasValue := splitFlag(arg)
		canonical, ok := fs.alias[namePart]
		if !ok {
			return nil, fmt.Errorf("未知选项: %s", arg)
		}

		if _, isString := fs.stringValues[canonical]; isString {
			if !hasValue {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("选项 --%s 需要一个值", namePart)
				}
				next := args[i+1]
				if strings.HasPrefix(next, "-") && next != "-" && next != "--" {
					return nil, fmt.Errorf("选项 --%s 需要一个值", namePart)
				}
				valuePart = next
				i += 2
			} else {
				i++
			}
			fs.stringValues[canonical] = valuePart
			fs.setFlags[canonical] = true
			continue
		}

		if _, isBool := fs.boolValues[canonical]; isBool {
			value := true
			if hasValue {
				parsed, err := strconv.ParseBool(valuePart)
				if err != nil {
					return nil, fmt.Errorf("选项 --%s 的布尔值无效: %s", namePart, valuePart)
				}
				value = parsed
			}
			fs.boolValues[canonical] = value
			fs.setFlags[canonical] = true
			i++
			continue
		}

		return nil, fmt.Errorf("未知选项: %s", arg)
	}

	return remaining, nil
}

func (fs *flagSet) getString(name string) string {
	if fs == nil {
		return ""
	}
	canonical := fs.canonicalName(name)
	if v, ok := fs.stringValues[canonical]; ok {
		return v
	}
	return ""
}

func (fs *flagSet) getBool(name string) bool {
	if fs == nil {
		return false
	}
	canonical := fs.canonicalName(name)
	if v, ok := fs.boolValues[canonical]; ok {
		return v
	}
	return false
}

func (fs *flagSet) isSet(name string) bool {
	if fs == nil {
		return false
	}
	canonical := fs.canonicalName(name)
	return fs.setFlags[canonical]
}

func (fs *flagSet) canonicalName(name string) string {
	if fs == nil {
		return strings.ToLower(name)
	}
	if canonical, ok := fs.alias[strings.ToLower(name)]; ok {
		return canonical
	}
	return strings.ToLower(name)
}

func splitFlag(arg string) (name, value string, hasValue bool) {
	trimmed := strings.TrimLeft(arg, "-")
	if idx := strings.Index(trimmed, "="); idx >= 0 {
		name = strings.ToLower(strings.TrimSpace(trimmed[:idx]))
		value = trimmed[idx+1:]
		hasValue = true
		return
	}
	name = strings.ToLower(strings.TrimSpace(trimmed))
	value = ""
	return
}

func parseFlagNames(name string) []string {
	parts := strings.Split(name, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		n := strings.ToLower(strings.TrimSpace(part))
		if n != "" {
			result = append(result, n)
		}
	}
	return result
}

func newContext(app *App, command *Command, args []string, fs *flagSet) *Context {
	return &Context{
		App:     app,
		Command: command,
		args:    args,
		flags:   fs,
	}
}
