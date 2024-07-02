//go:build cgo
// +build cgo

package native

/*
#if defined(__APPLE__) || defined(__linux__)
// https://github.com/wailsapp/wails/pull/2152/files#diff-d4a0fa73df7b0ab971e550f95249e358b634836e925ace96f7400480916ac09e
#include <errno.h>
#include <signal.h>
#include <stdio.h>
#include <string.h>

static void fix_signal(int signum)
{
    struct sigaction st;

    if (sigaction(signum, NULL, &st) < 0) {
        goto fix_signal_error;
    }
    st.sa_flags |= SA_ONSTACK;
    if (sigaction(signum, &st,  NULL) < 0) {
        goto fix_signal_error;
    }
    return;
fix_signal_error:
        fprintf(stderr, "error fixing handler for signal %d, please "
                "report this issue to "
                "https://github.com/pact-foundation/pact-go: %s\n",
                signum, strerror(errno));
}

static void install_signal_handlers()
{
#if defined(SIGCHLD)
    fix_signal(SIGCHLD);
#endif
#if defined(SIGHUP)
    fix_signal(SIGHUP);
#endif
#if defined(SIGINT)
    fix_signal(SIGINT);
#endif
#if defined(SIGQUIT)
    fix_signal(SIGQUIT);
#endif
#if defined(SIGABRT)
    fix_signal(SIGABRT);
#endif
#if defined(SIGFPE)
    fix_signal(SIGFPE);
#endif
#if defined(SIGTERM)
    fix_signal(SIGTERM);
#endif
#if defined(SIGBUS)
    fix_signal(SIGBUS);
#endif
#if defined(SIGSEGV)
    fix_signal(SIGSEGV);
#endif
#if defined(SIGXCPU)
    fix_signal(SIGXCPU);
#endif
#if defined(SIGXFSZ)
    fix_signal(SIGXFSZ);
#endif
}
#else
	static void install_signal_handlers()
	{
	}
#endif
*/
import "C"
import (
	"os"
	"runtime"
)

func InstallSignalHandlers() {
	if os.Getenv("PACT_GO_INSTALL_SIGNAL_HANDLERS") != "0" {
		if runtime.GOOS != "windows" {
			C.install_signal_handlers()
		}
	}
}
