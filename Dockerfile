# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

FROM golang as build

ADD . /notify/
WORKDIR /notify
RUN go mod download
RUN go build ./cmd/notify-api

#####

FROM ronmi/mingo

COPY --from="build" /notify/notify-api /
ENTRYPOINT ["/notify-api"]
