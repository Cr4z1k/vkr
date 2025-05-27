package parser

const (
	configDir = "/config"

	defaultConfigStartingString = `pipeline:
  processors:
`

	defaultJoinConfigTemplate = `pipeline:
  processors:
    # Добавляем мету: из какого топика пришло сообщение и какой ключ
    - mapping: |
        root = this
        meta join_key = this.%[1]s
        meta prev_missing = "false"
        meta body = this.string()

    - log:
        level: INFO
        message: |-
          after META ADDING
          this: ${! json() }
          content(): ${! content().string() }

    # Пробуем вытащить "парное" сообщение из кэша
    - cache:
        resource: join_cache
        operator: get
        key: ${! meta("join_key") }

    - log:
        level: INFO
        message: |-
          after CACHE GET
          this: ${! json() }
          content(): ${! content().string() }

    # Если не нашли — кладём своё и не пишем в output (catch)
    - catch:
        - mapping: meta prev_missing = "true"

    # Сохраняем в кэш (если это первое по ключу)
    - switch:
        - check: meta("prev_missing").bool()
          processors:
            - cache:
                resource: join_cache
                operator: set
                key: ${! meta("join_key") }
                value: ${! content() }
            # После set не отправляем в output

        # Если оба сообщения есть — делаем merge и удаляем из кэша
        - check: '!meta("prev_missing").bool()'
          processors:
            - mapping: |-
                let other = this
                let curr  = meta("body").string().parse_json()
                # Слияние данных — что-то вроде:
                let merged = $curr.merge($other)
                root = $merged
                root.%[1]s = meta("join_key")
            - cache:
                resource: join_cache
                operator: delete
                key: ${! meta("join_key") }

    - log:
        level: INFO
        message: |-
          after SWITCH
          this: ${! json() }
          content(): ${! content().string() }`

	joinFilterCondition = `
    - mapping: |
        root = if %s {
          this
        } else {
          deleted()
        }`
)
