import "source-map-support/register";

import { connect } from "mongoose";
import { InsertLog } from "./logs/logs.services.js";
import {
  Server,
  ServerCredentials,
  ServerUnaryCall,
  sendUnaryData,
  ServerErrorResponse,
  UntypedHandleCall,
} from "@grpc/grpc-js";
import {
  LogRequest,
  LogServiceServer,
  LogResponse,
  LogServiceService,
} from "../protos/logs.js";

async function connectToMongo() {
  try {
    await connect("mongodb://127.0.0.1:27017/");
  } catch (err) {
    console.log(err);
  }
}

class Logger implements LogServiceServer {
  [method: string]: UntypedHandleCall;

  public async insertLog(
    call: ServerUnaryCall<LogRequest, LogResponse>,
    callback: sendUnaryData<LogResponse>
  ) {
    console.log("Comming Request!");
    const res: Partial<LogResponse> = {};

    const log = call.request.log;

    if (typeof log === "undefined") {
      const error: ServerErrorResponse = {
        name: "Log Missing",
        message: "log request body is empty",
      };
      console.log("Faild Request!");
      callback(error, null);
      return;
    }

    try {
      await InsertLog({
        service_name: log.serviceName,
        error_message: log.errorMessage,
        stack_trace: log.stackTrace,
      });
    } catch (err: any) {
      const error: ServerErrorResponse = {
        name: "Unexpcted Error",
        message: err.message,
      };
      console.log("Faild2 Request!");
      callback(error, null);
      return;
    }
    console.log("Successful Request!");
    res.message = "log was inserted successfuly";
    callback(null, LogResponse.fromJSON(res));
  }
}

const server = new Server({
  "grpc.max_receive_message_length": -1,
  "grpc.max_send_message_length": -1,
});

server.addService(LogServiceService, new Logger());
const startServer = () => {
  server.bindAsync("0.0.0.0:50051", ServerCredentials.createInsecure(), () => {
    server.start();

    console.log("logger Service is running on 0.0.0.0:50051");
  });
};
(async () => {
  try {
    await connectToMongo();
    startServer();
  } catch (err) {
    console.error(err);
  }
})();
