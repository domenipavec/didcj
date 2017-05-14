#include <assert.h>
#include <fstream>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>

typedef struct buffer {
	char* buf;
	int size;
	int pos; // for input buffers next byte to be read. for output buffers next byte to be written.
} buffer;

static const int MAX_MACHINES = 100;
static const int MAX_MESSAGE_SIZE = (8 * (1 << 20));
static const char SEND = 0;
static const char RECEIVE = 1;
static const char DEBUG = 2;
static buffer incoming_buffers[MAX_MACHINES];
static buffer outgoing_buffers[MAX_MACHINES];

int NumberOfNodes() {
	return %d;
}

void fputint(int value, FILE * out) {
	for (int i = 0; i < 4; i++) {
		fputc((0xff & (value >> (8*i))), out);
	}
}

void Debug(const char *s) {
	fputc(DEBUG, stderr);
	fputint(strlen(s), stderr);
	fputs(s, stderr);
}

static void putRawByte(buffer* buf, unsigned char byte) {
	if (buf->pos >= buf->size) {
		buf->size = 2 * buf->size;
		if (buf->size < 128) buf->size = 128;
		buf->buf = (char*)realloc(buf->buf, buf->size);
		assert(buf->buf);
	}
	buf->buf[buf->pos++] = byte;
}

static inline void die(const char *s) {
	Debug(s);
	exit(20);
}

static inline void checkNodeId(int node) {
	if (node < 0 || node >= NumberOfNodes()) {
		die("Incorrect node number!");
	}
}

int MyNodeId() {
	static int id = -1;
	if (id == -1) {
		std::fstream node_id_file("nodeid", std::ios_base::in);
		node_id_file >> id;
	}
	return id;
}

void PutChar(int target, char value) {
	checkNodeId(target);
	putRawByte(&outgoing_buffers[target], value);
}

void PutInt(int target, int value) {
	checkNodeId(target);
	for (int i = 0; i < 4; i++) {
		putRawByte(&outgoing_buffers[target], (0xff & (value >> (8*i))));
	}
}

void PutLL(int target, long long value) {
	checkNodeId(target);
	for (int i = 0; i < 8; i++) {
		putRawByte(&outgoing_buffers[target], (0xff & (value >> (8*i))));
	}
}

void Send(int target) {
	checkNodeId(target);

	buffer *buf = &outgoing_buffers[target];
	if (buf->pos > MAX_MESSAGE_SIZE) {
		die("Send message too long!");
	}

	fputc(SEND, stderr);
	fputint(target, stderr);
	fputint(buf->pos, stderr);
	fwrite(buf->buf, sizeof(char), buf->pos, stderr);
	fflush(stderr);

	free(buf->buf);
	buf->buf = NULL;
	buf->pos = 0;
	buf->size = 0;
}

void freadint(int *value, FILE * in) {
	while (fread(value, 4, 1, in) < 1);
}

int Receive(int source) {
	checkNodeId(source);

	fputc(RECEIVE, stderr);
	fputint(source, stderr);
	fflush(stderr);

	int length;
	freadint(&length, stdin);
	int sender;
	freadint(&sender, stdin);
	assert(length <= MAX_MESSAGE_SIZE);

	buffer *buf = &incoming_buffers[sender];
	if (buf->buf != NULL) {
		free(buf->buf);
		buf->buf = NULL;
	}
	buf->buf = (char *)malloc(length);
	assert(buf->buf);

	while (fread(buf->buf, length, 1, stdin) < 1);
	buf->pos = 0;
	buf->size = length;

	return sender;
}

static unsigned char getRawByte(buffer *buf) {
	if (buf->pos >= buf->size) {
		die("Read past the end of msg!");
	}
	char r = buf->buf[buf->pos++];
	if (buf->pos >= buf->size) {
		free(buf->buf);
		buf->buf = NULL;
		buf->pos = 0;
		buf->size = 0;
	}
	return r;
}

char GetChar(int source) {
	checkNodeId(source);

	return getRawByte(&incoming_buffers[source]);
}

int GetInt(int source) {
	checkNodeId(source);

	int result = 0;
	for (int i = 0; i < 4; i++) {
		result |= (int)(getRawByte(&incoming_buffers[source])) << (8 * i);
	}
	return result;
}

long long GetLL(int source) {
	checkNodeId(source);

	long long result = 0;
	for (int i = 0; i < 8; i++) {
		result |= (long long)(getRawByte(&incoming_buffers[source])) << (8 * i);
	}
	return result;
}
