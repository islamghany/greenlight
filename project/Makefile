        



  # listener-service:
  #   build:
  #     context: ./../listener-service
  #     dockerfile: ./../listener-service/Dockerfile
  #   restart: always
  #   networks:
  #     greenlight-net:
  # mail-service:
  #   build:
  #     context: ./../mail-service
  #     dockerfile: ./../mail-service/Dockerfile
  #   restart: always
  #   environment:
  #     MAIL_DOMAIN: localhost
  #     MAIL_HOST: mailhog
  #     MAIL_PORT: 1025
  #     MAIL_ENCRYPTION: none
  #     MAIL_USERNAME: ""
  #     MAIL_PASSWORD: ""
  #     FROM_NAME: "islam ghany"
  #     FROM_ADDRESS: islam.ghany@example.com
  #   networks:
  #     greenlight-net:


  


  
        # mailhog:
        #   image: "mailhog/mailhog:latest"
        #   restart: always
        #   ports:
        #     - "1025:1025"
        #     - "8025:8025"
        # networks:
        #   greenlight-net:


        # rabbitmq:
        #   image: "rabbitmq:3.10.7-alpine"
        #   ports:
        #     - "5672:5672"
        #   deploy:
        #     mode: replicated
        #     replicas: 1
        #   volumes:
        #     - ./db-data/rabbitmq/:/var/lib/rabbitmq/
        #   networks:
        #     greenlight-net: