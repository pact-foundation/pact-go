package verifier

/*
/*
#cgo LDFLAGS: ${SRCDIR}/../../../../libs/libpact_verifier_cli.so

// Library headers
typedef int bool;
#define true 1
#define false 0

void init(char* log);
char* version();
void free_string(char* s);
int verify(char* s);

*/
import "C"
