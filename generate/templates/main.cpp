#include <message.h>
#include "%s.h"

#include <algorithm>
#include <iostream>
#include <map>
#include <set>
#include <stdint.h>
#include <vector>

using namespace std;


int main() {
    int64_t TOTAL_N = %s();
    int64_t N = (TOTAL_N + NumberOfNodes() - 1)/NumberOfNodes();
    int64_t nodes = (TOTAL_N + N - 1)/N;
    if (MyNodeId() >= nodes) {
        return 0;
    }

    int64_t start = N*MyNodeId();
    int64_t end = N*(MyNodeId() + 1);
    if (end > TOTAL_N) {
        end = TOTAL_N;
    }

    int64_t result = 0;
    for (int64_t i = start; i < end; i++) {

    }

    PutLL(0, result);
    Send(0);

    result = 0;
    if (MyNodeId() == 0) {
        for (int64_t i = 0; i < nodes; i++) {
            Receive(i);
            result += GetLL(i);
        }

        cout << result << endl;
    }

    return 0;
}
