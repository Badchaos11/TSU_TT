# TSU_TT


## Общее описание
API создания, изменения и просмотра пользователей. Роуты реализованы при помощи gorilla/mux для большего удобства объединения эндпоинтов. База данных использована PostgreSQL, движок базы данных pgx. Кэширование входящих запросов на поиск выполняется с помощбю Redis, раз в час выполняется очистка кэша. Конфигурация загружается из .env файла, для развертывания в Docker предусмотрен отдельный файл. Логирование выполняется библиотекой logrus.

## Эндпоинты

**/create_user** - Создание пользователя, метод POST. На вход поступает  json объект, пример: {"name": "andrei", "surname": "yurchenko", "patronymic": "mikhailovich", "sex": "male", "birth_date": "1998-02-08T10:15:12.822415Z"}. Все поля строковые, кроме birth_date, patronymic и birth_date не обязательны к вводу. В случае успешного создания пользователя будет сообщено id записи.

**/change_user** - Изменение пользователя, метод POST. На вход поступает json обЪект, пример {"id" 1, "name": "t", "surname": "t", patronymic": "t", "sex": "female", "status": "blocked", "birth_date": "t"}. Все поля кроме id и birht_date строковые. ID обязателен к вводу, остальные нет. В случае если пользователь не будет найден, возвращается код 404, так как пользователь не был найден. В случае успеха будет доставлено сообщение об этом.

**/delete_user** - Удаление пользователя, метод DELETE. На вход поступает json объект, пример {"user_id": 1}. Если пользователь не был найден, возвращается 404 код и сообщение об ошибке. В случае успеха сообщение об этом.

**/create_user_from_file** - Создание пользователя из файла, метод POST. В формате multipart/form передается xlsx\xls файл под ключом user. Если файл удается принять и прочитать, создается пользователь по данным из него. Данные в файле должны располагаться в первой колонке А на строках 1-5, в порядке Имя, Фамилия, Отчество, Пол, Почта. В случае успеха возвращается id записи. Почта считывается из файла, выводится в логе, он не используется.

**/get_user_by_id** - Получение одного пользователя по его id, метод GET. Параметр user_id передается в секции query, пример localhost:3000/get_user_by_id?user_id=1. Если user_id отрицательный или содержит не только цифры, возвращается код 400 и сообщение о неправильно введенном id. Если пользователь не был найден в базе, возвращается код 404 и сообщение об этом. В случае успеха пользователь возвращается в json объекте

**/get_filtered_users** - Получение списка пользователей по заданному фильтру, метод GET. Все параметры передаются в секции query. Доступны следующие поля для фильтрации:
    - sex - выбор пола;
    - status - выбор статуса;
    - name - поиск по имени пользователя;
    - surname - поиск по фамилии;
    - patronymic - поиск по отчеству;
    - order_by - по какому атрибуту сортировать, sex или status;
    - desc - при наличии сортировки выбирается порядок: true - в порядке убывания, false - в порядке возрастания;
    - limit - количество получаемых результатов;
    - offset - сдвиг от начала полученных результатов.
В случае если не удалось найти пользователей по заданному фильтру, возвращается 404 код ошибки. Если нашёлся хоть 1 результат, возвращается массив json объектов. Если ни один из параметров не заполнен, будет выведен весь список пользователей.


## Кеширование

Кешируются запросы на поиск. При первом запросе в redis добавляется результат успешного поиска по соответствующему ключу. При повторном поиске будет проведена проверка наличия этой записи. В случае успеха возвращается запись из redis. Если не удалось найти запись, выполняется запрос в базу данных и добавление записи в кэш.


## Развертывание

Для развертывания представлены Dockerfile и docker-compose. Разворачиваются api, postgres, redis. В файле .env находятся параметры, которые используются для конфигурации сервисов. В файле local.env находятся параметры, используемые для развертывания на локальном устройстве.