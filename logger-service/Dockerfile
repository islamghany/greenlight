FROM node:19.2-alpine

WORKDIR /usr

COPY package.json ./

COPY dist ./dist

COPY wait-for.sh .
COPY start.sh .
RUN ls -a

RUN chmod +x wait-for.sh
RUN chmod +x start.sh

RUN npm install

EXPOSE 50051

CMD [ "npm", "run", "start" ]
