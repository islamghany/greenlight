/* eslint-disable */
import {
  CallOptions,
  ChannelCredentials,
  ChannelOptions,
  Client,
  ClientUnaryCall,
  handleUnaryCall,
  makeGenericClientConstructor,
  Metadata,
  ServiceError,
  UntypedServiceImplementation,
} from "@grpc/grpc-js";
import _m0 from "protobufjs/minimal.js";

export interface Log {
  serviceName: string;
  errorMessage: string;
  stackTrace: string;
}

export interface LogRequest {
  log?: Log;
}

export interface LogResponse {
  message: string;
}

function createBaseLog(): Log {
  return { serviceName: "", errorMessage: "", stackTrace: "" };
}

export const Log = {
  encode(message: Log, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.serviceName !== "") {
      writer.uint32(10).string(message.serviceName);
    }
    if (message.errorMessage !== "") {
      writer.uint32(18).string(message.errorMessage);
    }
    if (message.stackTrace !== "") {
      writer.uint32(26).string(message.stackTrace);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Log {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLog();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.serviceName = reader.string();
          break;
        case 2:
          message.errorMessage = reader.string();
          break;
        case 3:
          message.stackTrace = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): Log {
    return {
      serviceName: isSet(object.serviceName) ? String(object.serviceName) : "",
      errorMessage: isSet(object.errorMessage)
        ? String(object.errorMessage)
        : "",
      stackTrace: isSet(object.stackTrace) ? String(object.stackTrace) : "",
    };
  },

  toJSON(message: Log): unknown {
    const obj: any = {};
    message.serviceName !== undefined &&
      (obj.serviceName = message.serviceName);
    message.errorMessage !== undefined &&
      (obj.errorMessage = message.errorMessage);
    message.stackTrace !== undefined && (obj.stackTrace = message.stackTrace);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<Log>, I>>(object: I): Log {
    const message = createBaseLog();
    message.serviceName = object.serviceName ?? "";
    message.errorMessage = object.errorMessage ?? "";
    message.stackTrace = object.stackTrace ?? "";
    return message;
  },
};

function createBaseLogRequest(): LogRequest {
  return { log: undefined };
}

export const LogRequest = {
  encode(
    message: LogRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.log !== undefined) {
      Log.encode(message.log, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LogRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLogRequest();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.log = Log.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): LogRequest {
    return { log: isSet(object.log) ? Log.fromJSON(object.log) : undefined };
  },

  toJSON(message: LogRequest): unknown {
    const obj: any = {};
    message.log !== undefined &&
      (obj.log = message.log ? Log.toJSON(message.log) : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<LogRequest>, I>>(
    object: I
  ): LogRequest {
    const message = createBaseLogRequest();
    message.log =
      object.log !== undefined && object.log !== null
        ? Log.fromPartial(object.log)
        : undefined;
    return message;
  },
};

function createBaseLogResponse(): LogResponse {
  return { message: "" };
}

export const LogResponse = {
  encode(
    message: LogResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.message !== "") {
      writer.uint32(10).string(message.message);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): LogResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseLogResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.message = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): LogResponse {
    return { message: isSet(object.message) ? String(object.message) : "" };
  },

  toJSON(message: LogResponse): unknown {
    const obj: any = {};
    message.message !== undefined && (obj.message = message.message);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<LogResponse>, I>>(
    object: I
  ): LogResponse {
    const message = createBaseLogResponse();
    message.message = object.message ?? "";
    return message;
  },
};

export type LogServiceService = typeof LogServiceService;
export const LogServiceService = {
  insertLog: {
    path: "/logpb.LogService/InsertLog",
    requestStream: false,
    responseStream: false,
    requestSerialize: (value: LogRequest) =>
      Buffer.from(LogRequest.encode(value).finish()),
    requestDeserialize: (value: Buffer) => LogRequest.decode(value),
    responseSerialize: (value: LogResponse) =>
      Buffer.from(LogResponse.encode(value).finish()),
    responseDeserialize: (value: Buffer) => LogResponse.decode(value),
  },
} as const;

export interface LogServiceServer extends UntypedServiceImplementation {
  insertLog: handleUnaryCall<LogRequest, LogResponse>;
}

export interface LogServiceClient extends Client {
  insertLog(
    request: LogRequest,
    callback: (error: ServiceError | null, response: LogResponse) => void
  ): ClientUnaryCall;
  insertLog(
    request: LogRequest,
    metadata: Metadata,
    callback: (error: ServiceError | null, response: LogResponse) => void
  ): ClientUnaryCall;
  insertLog(
    request: LogRequest,
    metadata: Metadata,
    options: Partial<CallOptions>,
    callback: (error: ServiceError | null, response: LogResponse) => void
  ): ClientUnaryCall;
}

export const LogServiceClient = makeGenericClientConstructor(
  LogServiceService,
  "logpb.LogService"
) as unknown as {
  new (
    address: string,
    credentials: ChannelCredentials,
    options?: Partial<ChannelOptions>
  ): LogServiceClient;
  service: typeof LogServiceService;
};

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;

type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & {
      [K in Exclude<keyof I, KeysOfUnion<P>>]: never;
    };

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
