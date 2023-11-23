vcl 4.1;

import directors;

# Default backend definition. Set this to point to your content server.
# backend default {
#     .host = "127.0.0.1";
#     .port = "8080";
# }

backend isu1 {
    .host = "192.168.0.11";
    .port = "8080";
}

backend isu2 {
    .host = "192.168.0.12";
    .port = "8080";
}

# backend isu3 {
#     .host = "192.168.0.13";
#     .port = "8080";
# }

sub vcl_init {
    # ラウンドロビンでリクエストを送る
    new bar = directors.round_robin();
    bar.add_backend(isu1);
    bar.add_backend(isu2);
}

sub vcl_recv {
    # 特定パスだけは別のバックエンドに送る
    if (req.url ~ "^/java/") {
        set req.backend_hint = isu1;
    } else {
        set req.backend_hint = isu2;
    }

    # 特定のURLをキャッシュしないようにする
    if (req.url ~ "^/nocache") {
        return (pass);
    }

    # クッキーを削除してキャッシュを可能にする
    if (req.url ~ "^/cacheable") {
        unset req.http.cookie;
    }
}

sub vcl_backend_response {
    # レスポンスのキャッシュ時間を設定
    set beresp.ttl = 0.5s;

    # 特定のパスに対するTTLの設定
    if (bereq.url ~ "^/short-term") {
        set beresp.ttl = 0.5s;
    } elseif (bereq.url ~ "^/long-term") {
        set beresp.ttl = 10m;
    } else {
        set beresp.ttl = 1s;
    }

    # キャッシュされたオブジェクトの有効期限後も、一定時間レスポンスを提供する
    # この間にバックエンドからのレスポンスが返ってきたら、それをキャッシュする
    # stale-while-revalidate を実現する
    set beresp.grace = 1s;

    # キャッシュされたオブジェクトの有効期限後に、varnishがキャッシュを保持する時間
    # If-Modified-Since: and/or Ìf-None-Match がある場合に 304 Not Modifiedを実現するために設定
    set beresp.keep = 30s;

    # varnishでgzipしたい場合は以下を追加
    if (beresp.http.content-type ~ "text" || beresp.http.content-type ~ "json" ) {
        set beresp.do_gzip = true;
    }
}


sub vcl_deliver {
    # レスポンスヘッダーにキャッシュの状態を追加
    if (obj.hits > 0) {
        set resp.http.X-Cache = "HIT";
    } else {
        set resp.http.X-Cache = "MISS";
    }

    # 回によってはCache-Control: privateを付与しないとfailすることがあるので追加。基本varnish以降cdnを挟まないのでつけておいてもデメリットはない。
    # 例. https://github.com/isucon/isucon12-qualify/blob/main/webapp/specification.md#response%E3%81%AEhttp-header
    if (resp.http.Cache-Control) {
        set resp.http.Cache-Control = resp.http.Cache-Control + ", private";
    } else {
        set resp.http.Cache-Control = "private";
    }
}
