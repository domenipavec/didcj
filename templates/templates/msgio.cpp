#include <memory>
#include <stdint.h>
#include <vector>

void put(const int32_t target, const int8_t value) {
    PutChar(target, value);
}

void put(const int32_t target, const int32_t value) {
    PutInt(target, value);
}

void put(const int32_t target, const int64_t value) {
    PutLL(target, value);
}

template<typename T> void put(const int32_t target, const T& object) {
    const int8_t* begin = reinterpret_cast<const int8_t*>(std::addressof(object));

    for (int64_t i = 0; i < sizeof(T); i++) {
        put(target, *(begin + i));
    }
}

template<typename T> void put(const int32_t target, const std::vector<T>& object) {
    int64_t size = object.size();
    put(target, size);
    for (int64_t i = 0; i < size; i++) {
        put(target, object[i]);
    }
}

void get(const int32_t target, int8_t& value) {
    value = GetChar(target);
}

void get(const int32_t target, int32_t& value) {
    value = GetInt(target);
}

void get(const int32_t target, int64_t& value) {
    value = GetLL(target);
}

template<typename T> void get(const int32_t target, T& object) {
    int8_t* begin = reinterpret_cast<int8_t*>(std::addressof(object));

    for (int64_t i = 0; i < sizeof(T); i++) {
        get(target, *(begin + i));
    }
}

template<typename T> void get(const int32_t target, std::vector<T>& object) {
    int64_t size;
    get(target, size);
    object.clear();
    object.reserve(size);
    for (int64_t i = 0; i < size; i++) {
        T obj;
        get(target, obj);
        object.push_back(obj);
    }
}
