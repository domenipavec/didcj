#include <stdint.h>
#include <iostream>

class Mod {
protected:
    int64_t value;

    void fix() {
        if (value >= PRIME || value < 0) {
            value = value%PRIME;
            if (value < 0) {
                value += PRIME;
            }
        }
    }

public:
    Mod(int64_t v = 0) {
        value = v;
        fix();
    }

    Mod& operator++() {
        value++;
        if (value >= PRIME) {
            value -= PRIME;
        }
        return *this;
    }

    Mod operator++(int) {
        Mod tmp(*this);
        operator++();
        return tmp;
    }

    Mod& operator--() {
        value--;
        if (value < 0) {
            value += PRIME;
        }
        return *this;
    }

    Mod operator--(int) {
        Mod tmp(*this);
        operator--();
        return tmp;
    }

    Mod& operator+=(const Mod& rhs) {
        value += rhs.value;
        if (value >= PRIME) {
            value -= PRIME;
        }
        return *this;
    }

    Mod& operator-=(const Mod& rhs) {
        value -= rhs.value;
        if (value < 0) {
            value += PRIME;
        }
        return *this;
    }

    Mod& operator*=(const Mod& rhs) {
        value *= rhs.value;
        fix();
        return *this;
    }

    Mod& operator/=(const Mod& rhs) {
        return operator*=(exp(rhs, PRIME-2));
    }

    Mod& operator^=(const int64_t& rhs) {
        if (rhs < 0) {
            *this = exp(exp(value, -1*rhs), PRIME-2);
        } else if (rhs == 0) {
            value = 1;
        } else if (rhs == 1) {

        } else if (rhs % 2 == 0) {
            *this = exp(value*value, rhs/2);
        } else {
            *this = value*exp(value*value, (rhs-1)/2);
        }
        return *this;
    }

    int64_t Get() const {
        return value;
    }

    friend bool operator==(const Mod& a, const Mod& b) {
        return a.Get() == b.Get();
    }

    friend std::ostream& operator<<(std::ostream& out, const Mod& m) {
        return out << m.Get();
    }

    friend Mod operator+(Mod lhs, const Mod& rhs) {
        lhs += rhs;
        return lhs;
    }

    friend Mod operator-(Mod lhs, const Mod& rhs) {
        lhs -= rhs;
        return lhs;
    }

    friend Mod operator*(Mod lhs, const Mod& rhs) {
        lhs *= rhs;
        return lhs;
    }

    friend Mod exp(Mod lhs, const int64_t& rhs) {
        lhs ^= rhs;
        return lhs;
    }

    friend Mod operator/(Mod lhs, const Mod& rhs) {
        lhs /= rhs;
        return lhs;
    }
};
