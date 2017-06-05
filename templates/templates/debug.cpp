#ifdef DIDCJ

void Timer(const char *s) {
	fputc(TIMER, stderr);
	fputint(strlen(s), stderr);
	fputs(s, stderr);
}

#else

#define Timer(s)
#define Debug(s)

#endif
