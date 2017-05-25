FROM ubuntu:16.04

RUN apt-get update
RUN apt-get install -y openssh-server psmisc locales

RUN locale-gen en_US.UTF-8
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8

RUN mkdir /var/run/sshd
RUN mkdir /root/.ssh
RUN echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCkWvR3axrI21rkyqofrl28AAtXastCqDU97CRkKwt7FBunNUiOdxo/e/hldgxqDLLWeqTQ9IFcTPHbhFXjEc7tateTeu5KIM/DA0c6mh4FMigTBABlN4h0OHX8qH5WkoisegRa0NeYGOhsgYw3QnTO32/dFFKQyVSPOZLuzsoCp5W5HV+5LEdnPgmeLwiSrz1ZIBfHAluTqgC+i91GGww0xAyIlUMMzw92csHTHxULwjpv3kAHwdntw57pdp1cCakOfl88mPdYR/fT5B9O4Mgr7ApXyxWWMb7Zd+5a5p2fNLjmUWWPIVf8NgCKKf9miRivDMf5L/7T3kZoHEfdpGc0wXWdGEsm4kMdI0N7pB0Vjqkm59mD9f4FX4FcL0HITKoYs+bCEbDzTNXUeTd/7hOmS8hbd0QdmOWnqqF1cMl6byBQ8lw6yi2nuUFHl9xK5+4/pOZmwlz6FD/G10ajOWcC5o8ZOn8lLNGqN+M7ElwQU6omX8jT0HeLd/U7yYqUZ/+6bkA/nb8I9OurKPqi3uhSpgLhTf7GPEufDtt2BdXpNdJB0V6uCD7fKJR+uyNA3sD+IOFaLswai37GduX6CQa5z+XYupfVX0fBvzFCtYsJZwfl9tZCI5HY6+kN4YqzjOdU382gT+8Tgt9TUR7XMjfYRJCR+t++GjUNO4gbLpr2aQ== domen@desktop" >> /root/.ssh/authorized_keys

CMD ["/usr/sbin/sshd", "-D"]
