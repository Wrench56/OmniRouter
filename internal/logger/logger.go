package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	defaultFilePath   = "logs/app.log"
	defaultTimeFormat = time.RFC3339
	defaultMaxSizeMB  = 64
	defaultMaxBackups = 5
	defaultMaxAgeDays = 30

	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"

	ansiGray    = "\x1b[37m"
	ansiCyan    = "\x1b[36m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiRed     = "\x1b[31m"
	ansiMagenta = "\x1b[35m"

	callerWidth = 16
)

var (
	defaultLevel = zerolog.InfoLevel
	levelToken   = map[string]string{
		"trace":   ansiBold + ansiGray + "[TRC]" + ansiReset,
		"debug":   ansiBold + ansiCyan + "[DBG]" + ansiReset,
		"info":    ansiBold + ansiGreen + "[INF]" + ansiReset,
		"warn":    ansiBold + ansiYellow + "[WRN]" + ansiReset,
		"warning": ansiBold + ansiYellow + "[WRN]" + ansiReset,
		"error":   ansiBold + ansiRed + "[ERR]" + ansiReset,
		"fatal":   ansiBold + ansiMagenta + "[FTL]" + ansiReset,
		"panic":   ansiBold + ansiMagenta + "[PNC]" + ansiReset,
	}
)

func Setup() {
	fileWriter := &lumberjack.Logger{
		Filename:   defaultFilePath,
		MaxSize:    defaultMaxSizeMB,
		MaxBackups: defaultMaxBackups,
		MaxAge:     defaultMaxAgeDays,
		Compress:   true,
	}

	isTTY := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())

	var console io.Writer = os.Stderr
	if isTTY {
		cw := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: defaultTimeFormat,
			NoColor:    false,
		}

		cw.FormatLevel = func(i any) string {
			s := strings.ToLower(fmt.Sprint(i))
			if cw.NoColor {
				if len(s) > 3 {
					s = s[:3]
				}
				return "[" + strings.ToUpper(s) + "]"
			}
			if tok, ok := levelToken[s]; ok {
				return tok
			}
			up := strings.ToUpper(s)
			if len(up) > 3 {
				up = up[:3]
			}
			return ansiBold + ansiGray + "[" + up + "]" + ansiReset
		}

		cw.FormatCaller = func(i any) string {
			s, ok := i.(string)
			if !ok {
				s = fmt.Sprint(i)
			}

			if c := strings.IndexByte(s, ':'); c > 0 && c >= 3 && s[c-3:c] == ".go" {
				s = s[:c-3] + s[c:]
			}

			if callerWidth > 0 && len(s) > callerWidth {
				s = "~" + s[len(s)-callerWidth+1:]
			}

			return fmt.Sprintf("%-*s", callerWidth, s)
		}

		cw.PartsOrder = []string{
			zerolog.TimestampFieldName,
			zerolog.CallerFieldName,
			zerolog.LevelFieldName,
			zerolog.MessageFieldName,
		}

		console = cw
	}

	multi := zerolog.MultiLevelWriter(console, fileWriter)

	zerolog.TimeFieldFormat = defaultTimeFormat
	zerolog.SetGlobalLevel(defaultLevel)

	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		if IsLogCallerModuleSet() {
			return ConsumeLogCallerModule()
		}

		if idx := strings.LastIndexAny(file, `/\`); idx >= 0 {
			file = file[idx+1:]
		}
		return fmt.Sprintf("%s:%d", file, line)
	}

	/* Increase frame skips due to the wrapper we have */
	zerolog.CallerSkipFrameCount = 3
	log.Logger = zerolog.New(multi).
		With().
		Timestamp().
		Caller().
		Logger()
}

func SetLevel(level zerolog.Level) { defaultLevel = level; zerolog.SetGlobalLevel(level) }

func With(fields map[string]any) zerolog.Logger {
	return log.Logger.With().Fields(fields).Logger()
}

func Debug(msg string, kv ...any) { addKV(log.Debug(), kv...).Msg(msg) }
func Info(msg string, kv ...any)  { addKV(log.Info(), kv...).Msg(msg) }
func Warn(msg string, kv ...any)  { addKV(log.Warn(), kv...).Msg(msg) }
func Error(msg string, kv ...any) { addKV(log.Error(), kv...).Msg(msg) }
func Fatal(msg string, kv ...any) { addKV(log.Fatal(), kv...).Msg(msg) }

func DebugContext(ctx context.Context, msg string, kv ...any) {
	addKV(log.Ctx(ctx).Debug(), kv...).Msg(msg)
}
func InfoContext(ctx context.Context, msg string, kv ...any) {
	addKV(log.Ctx(ctx).Info(), kv...).Msg(msg)
}
func WarnContext(ctx context.Context, msg string, kv ...any) {
	addKV(log.Ctx(ctx).Warn(), kv...).Msg(msg)
}
func ErrorContext(ctx context.Context, msg string, kv ...any) {
	addKV(log.Ctx(ctx).Error(), kv...).Msg(msg)
}
func FatalContext(ctx context.Context, msg string, kv ...any) {
	addKV(log.Ctx(ctx).Fatal(), kv...).Msg(msg)
}

func addKV(e *zerolog.Event, kv ...any) *zerolog.Event {
	if len(kv)%2 != 0 {
		e = e.Interface("_kv_error", "odd number of keyvals")
	}
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			e = e.Interface(k, kv[i+1])
		} else {
			e = e.Interface("_kv_error", "non-string key")
		}
	}
	return e
}
