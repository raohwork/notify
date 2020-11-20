<?php
/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

namespace RaohWork\Notify;

use Exception;

class NotifyClient
{
    private $host;

    public function __construct(string $h)
    {
        $this->host = $h;
    }

    public function send(string $id, string $ep, string $driver, $data, bool $once=false): bool
    {
        $cmd = $once?'sendOnce':'send';
        try {
            $this->call($cmd, [
                'id' => $id,
                'type' => $driver,
                'endpoint' => $ep,
                'payload' => $data,
            ]);

            return true;
        } catch(Exception $e) {
            return false;
        }
    }

    private function call(string $cmd, $data): string
    {
        $cmd = ltrim($cmd, '/');
        $param = [
            'http' => [
                'method' => 'POST',
                'header' => "Content-Type: application/json\r\n",
                'content' => json_encode($data, \JSON_UNESCAPED_UNICODE),
            ],
        ];

        $ctx = stream_context_create($param);
        $fp = fopen($this->host . '/' . $cmd, 'rb', false, $ctx);
        if (!$fp) {
            throw new Exception('cannot connect to remote');
        }
        $ret = stream_get_contents($fp);
        fclose($fp);

        return $ret;
    }

    public function httpget(string $id, string $uri, ?array $h, ?array $val, bool $once=false): bool
    {
        $data = ['headers' => null, 'values' => null];
        if (!empty($h)) {
            $data['headers'] = $h;
        }
        if (!empty($val)) {
            $data['values'] = $val;
        }

        return $this->send($id, $uri, 'HTTPGET', $data, $once);
    }

    public function httppost(string $id, string $uri, ?array $h, string $body, bool $once=false): bool
    {
        $data = ['body' => $body];
        if (!empty($h)) {
            $data['headers'] = $h;
        }

        return $this->send($id, $uri, 'HTTPPOST', $data, $once);
    }

    public function tgmd(string $id, string $chan, string $md, bool $once=false): bool
    {
        return $this->send($id, $chan, 'TGMarkdown', $md, $once);
    }

    // see https://godoc.org/github.com/sendgrid/sendgrid-go/helpers/mail#SGMailV3
    // for data structure of $options
    public function sendgrid(string $id, array $options, bool $once=false): bool {
        return $this->send($id, '', 'SENDGRID', $options, $once);
    }

    /**
     * Notify server should not resend this. Every8d.com takes responsibility to
     * resend SMS.
     *
     * @param $phone string +886987654321 or 0987654321, separates by comma
     * @param $time string time to send SMS (YmdHis)
     * @param $retry int valid time in minutes
     */
    public function smsav8d(string $id, string $phone, string $text, string $subject = '', string $time = '', int $retry = 0): bool
    {
        $data = [
            'content' => $text,
            'subject' => $subject,
            'time' => $time,
            'retry' => $retry,
        ];
        return $this->send($id, $phone, 'SMSAV8D', $data, true);
    }

    public function resend(string $id): bool
    {
        try {
            $this->call('resend', ['id' => $id]);
            return true;
        } catch(Exception $e) {
            return false;
        }
    }

    public function result(string $id): string
    {
        // base64
        return $this->call('result', ['id' => $id]);
    }

    public function status(string $id): array
    {
        return json_decode($this->call('status', ['id' => $id]), true);
    }

    public function detail(string $id): array
    {
        $ret = json_decode($this->call('detail', ['id' => $id]), true);
        $ret['content'] = base64_decode($ret['content']);
        if (!empty($ret['response'])) {
            $ret['response'] = base64_decode($ret['response']);
        }
        return $ret;
    }

    public function delete(string $id): bool
    {
        try {
            $this->call('delete', ['id' => $id]);
            return true;
        } catch(Exception $e) {
            return false;
        }
    }

    public function clear(int $ts): bool
    {
        try {
            $this->call('clear', ['before' => $ts]);
            return true;
        } catch(Exception $e) {
            return false;
        }
    }

    public function forceClear(int $ts): bool
    {
        try {
            $this->call('forceClear', ['before' => $ts]);
            return true;
        } catch(Exception $e) {
            return false;
        }
    }
}
