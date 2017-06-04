#include <stdint.h>

inline int64_t ceil_div(const int64_t a, const int64_t b) {
    return (a + b - 1) / b;
}

int64_t calculate_bounds(const int64_t total_nodes, const int64_t total_n, const int64_t node_id, int64_t *start, int64_t *end) {
    int64_t per_node = ceil_div(total_n, total_nodes);

    *start = per_node*node_id;
    *end = *start + per_node;
    if (*end > total_n) {
        *end = total_n;
    }
}
