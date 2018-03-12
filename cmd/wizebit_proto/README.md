## Setup

go version go1.9.2

System: Linux amd64.

## Install

### Построение приложения

Для работы приложения необходимо построить утилиту wizefs, для этого необходимо зайти в директорию `wizefs` и запустить `go build`.

GUI application is based on platform-native GUI library `andlabs/ui`, we have fork of this library `leedark/ui` and you should get it by `go get`:

```
go get -u github.com/leedark/ui
```

Данная библиотека основана на простой кросс-платформенной библиотеке, написанной на C: [andlabs/libui](https://github.com/andlabs/libui). Мы также сделали форк данной библиотеки для некоторых исправлений: [leedark/libui](https://github.com/LeeDark/libui).

**You must include this library in your binary distributions.**

Для ее построения необходим CMake 3.1.0 или новее.

Out-of-tree builds typical of cmake are preferred:

```
$ # you must be in the top-level libui directory, otherwise this won't work
$ mkdir build
$ cd build
$ cmake ..
```

Pass `-DBUILD_SHARED_LIBS=OFF` to `cmake` to build a static library. The standard cmake build configurations are provided; if none is specified, `Debug` is used.

If you use a makefile generator with cmake, then

```
$ make
$ make tester         # for the test program
$ make examples       # for examples
```

and pass `VERBOSE=1` to see build commands. Build targets will be in the `build/out` folder.

Then you should go to the directory `wizefs/proto` and run `go build`.

### Разворачивание и старт кластера



## Main Window

Приложение имеет две вкладки (Tab): Wallet и Storage. 

![wallet-tab](images/wallet-tab.png)

![storage-tab](images/storage-tab.png)



## Wallet Tab

На вкладке Wallet отображается информация о кошельке пользователя и список кошельков, полученных через WizeBlock API.

Для создания кошелька необходимо нажать кнопку Create Wallet. После чего кошелек будет создан и информация о нем обновится. Также обновится список кошельков, в который добавится только что созданный кошелек пользователя.

![wallet-tab-info](images/wallet-tab-info.png)



## Storage Tab

До создания кошелька пользователя фунциональность вкладки Storage недоступна.

Если кошелек пользователя уже создан, то мы можем выполнять действия со Storage: Put file, Get file и Remove file. Также будет доступен список файлов, загруженных пользователем.

![storage-tab-files](images/storage-tab-files.png)

Все действия с файлами записываются в журнал, расположенный под списком файлов.

## Примечание (для разработчиков)

Сохранение информации по кошельку сделано в файл wallet.json, чтобы было просто.

В приложении реализован опрос кластера (WizeBlock и Raft части) каждую минуту (можно установить другую длительность). ~~Пока не реализовано активирование GUI в случае запущенного кластера, и деактивирование GUI в случае незапущенного кластера.~~

## Задачи

1.  **Активирование и деактивирование GUI в случае запущенного и незапущеного кластера**
2.  *Улучшить GUI для Wallet Tab: отображение информации по кошельку, **получать данные по запросу /wallet/{hash} и вставлять их в список кошельков**
3.  *Реализовать BlockApi /send и отправлять транзакции
4.  ***Добавить в структуру хранилища Raft возможность шифрования Base64(File.Basename) для записи и дешифрования для чтения**
5.  Почистить и ускорить BlockApi и RaftApi клиентов
6.  Тестирование Put/Get file в случае больших файлов (вроде работает)
7.  Объединить команды Storage API: get и xget (указывается также полный путь с названием файла куда взять файл) в одну команду get
8.  Добавить алгоритм нормализации индекса CPKIndex в Raft при удалении файлов (а значит и записей о них)

Я пометил звездочкой (*) важные задачи.