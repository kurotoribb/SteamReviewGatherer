# SteamReviewGatherer
![image](https://user-images.githubusercontent.com/9584727/122040648-05169a80-ce13-11eb-951d-ae32b2873640.png)


## 概要
Steamのレビューを収集するアプリケーションです
対象としたタイトルの以下をcsvとして出力します
```
- タイトル
- 全世界のレビュー総数
- 全世界のポジティブレビュー
- 全世界のネガティブレビュー
- 日本語のレビュー総数
- 日本語のポジティブレビュー
- 日本語のネガティブレビュー
- 全世界のネガティブ評価比(レビュー総数/ネガティブ数)
- 日本語のネガティブ評価比(レビュー総数/ネガティブ数)
- レビューの日本語比(日本語レビュー総数/全世界のレビュー総数)
```

## 使い方
### 準備編
- 以下のライブラリに依存しています よしなにしてください
    - [gocarina/gocsv: The GoCSV package aims to provide easy CSV serialization and deserialization to the golang programming language](https://github.com/gocarina/gocsv)
    - [peppage/kettle: Steam API client written in golang](https://github.com/peppage/kettle)
- [Steam API Key](https://steamcommunity.com/dev/apikey)が必要になります 用意してください
- `appId.csv` というファイルが必要になるため、事前に調査対象のAppIdを収集してください

`appId.csv` は以下のようなフォーマットでバイナリと同じディレクトリへ配置してください
```appId.csv
1384160,
1517290,
1203220,
1426210,
1172620,
1517290,
1124300,
1328670,
1313860,
623280,
```

### 実行編
- Buildを作る、go runするなどお好きな方法で起動してください
- Steam API Keyを求められるので入力してください
- 処理が完了すると `sample.csv` が吐き出されます

## Licence
[The MIT License (MIT) | Open Source Initiative](https://opensource.org/licenses/mit-license.php)

## お願い
これを使用して議論の叩きとして使用することを歓迎します

ただ、その際には対象の選定や結果の表示はフェアに行ってください

また、このGitHub リポジトリへのリンクを書いてくれるとちょっと喜びます
