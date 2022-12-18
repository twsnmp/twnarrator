# twnarrator
This is a narrator using VOICEVOX.
VOICEVOXを使用したナレーターです。


## Overview

テキストファイルかパワーポイントのノートの書いた
台本を[VOICEVOX](https://voicevox.hiroshiba.jp)のエンジン
を使ってスライド毎の音声ファイルにします。

https://github.com/nobonobo/voicevox-cli

のコードを参考にしています。ありがとうございます。

## Status

実現したい機能は全て対応しました。

## Build

### Build Env
ビルドするためには、以下の環境が必要です。

- go 1.19

### Build
ビルドはmakeで行います。
以下のターゲットが指定できます。
```
  all        全実行ファイルのビルド（省略可能）
  mac        Mac用の実行ファイルのビルド
  clean      ビルドした実行ファイルの削除
  zip        リリース用のZIPファイルを作成
```

```
make
```
を実行すれば、MacOS,Windowsの実行ファイルが、`dist`のディレクトリに作成されます。
Linux版はソースコードからビルドしてください。
以下のライブラリが必要です。
```
apt install libasound2-dev
```

配布用のZIPファイルを作成するためには、
```
make zip
```
を実行します。ZIPファイルが`dist/`ディレクトリに作成されます。

## Run

Mac OS,Windows,Linuxの環境でコマンドを実行する場合は、
VOICEVOXのエンジンを駆動してください。
Docker版でもダウンロードしたGUI付きでもよいです。
Docker版は
```
docker run -d -p 50021:50021 hiroshiba/voicevox_engine:cpu-ubuntu20.04-0.10.4
```
で起動でます。

動作の確認のために

``` 
#./twnarrator -l
```
を実行するとスピーカーとスタイルのリストが表示されます。

``` 
#./twnarrator -s <台本のファイル>
```
です。
<台本のファイル>はテキスト形式、またはPower Pointファイル(pptx)に対応しています。
実行すると台本ファイルと同じディレクトリにスライド毎の音声ファイルが作成されます。

台本のファイルは

```
#玄野武宏,ノーマル
こんにちは、ひまりさん

#冥鳴ひまり,ノーマル,1.0,1.0,1.0,0.0
こんにちは、たけひろさん

$

#玄野武宏,ノーマル
お元気ですか？

```
のような形式です。

```
#冥鳴ひまり,ノーマル,1.0,1.0,1.0,0.0
```
のような行はスピーカー、スタイル、スピード、イントネーション、音量、ピッチ
を指定するものです。

テキスト版の台本では$でスライドの区切りになります。


## Copyright

see ./LICENSE

```
Copyright 2022 Masayuki Yamai
```
