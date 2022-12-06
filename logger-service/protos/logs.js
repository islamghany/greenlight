"use strict";
exports.__esModule = true;
exports.LogServiceClient = exports.LogServiceService = exports.LogResponse = exports.LogRequest = exports.Log = void 0;
/* eslint-disable */
var grpc_js_1 = require("@grpc/grpc-js");
var minimal_js_1 = require("protobufjs/minimal.js");
function createBaseLog() {
    return { serviceName: "", errorMessage: "", stackTrace: "" };
}
exports.Log = {
    encode: function (message, writer) {
        if (writer === void 0) { writer = minimal_js_1["default"].Writer.create(); }
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
    decode: function (input, length) {
        var reader = input instanceof minimal_js_1["default"].Reader ? input : new minimal_js_1["default"].Reader(input);
        var end = length === undefined ? reader.len : reader.pos + length;
        var message = createBaseLog();
        while (reader.pos < end) {
            var tag = reader.uint32();
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
    fromJSON: function (object) {
        return {
            serviceName: isSet(object.serviceName) ? String(object.serviceName) : "",
            errorMessage: isSet(object.errorMessage)
                ? String(object.errorMessage)
                : "",
            stackTrace: isSet(object.stackTrace) ? String(object.stackTrace) : ""
        };
    },
    toJSON: function (message) {
        var obj = {};
        message.serviceName !== undefined &&
            (obj.serviceName = message.serviceName);
        message.errorMessage !== undefined &&
            (obj.errorMessage = message.errorMessage);
        message.stackTrace !== undefined && (obj.stackTrace = message.stackTrace);
        return obj;
    },
    fromPartial: function (object) {
        var _a, _b, _c;
        var message = createBaseLog();
        message.serviceName = (_a = object.serviceName) !== null && _a !== void 0 ? _a : "";
        message.errorMessage = (_b = object.errorMessage) !== null && _b !== void 0 ? _b : "";
        message.stackTrace = (_c = object.stackTrace) !== null && _c !== void 0 ? _c : "";
        return message;
    }
};
function createBaseLogRequest() {
    return { log: undefined };
}
exports.LogRequest = {
    encode: function (message, writer) {
        if (writer === void 0) { writer = minimal_js_1["default"].Writer.create(); }
        if (message.log !== undefined) {
            exports.Log.encode(message.log, writer.uint32(10).fork()).ldelim();
        }
        return writer;
    },
    decode: function (input, length) {
        var reader = input instanceof minimal_js_1["default"].Reader ? input : new minimal_js_1["default"].Reader(input);
        var end = length === undefined ? reader.len : reader.pos + length;
        var message = createBaseLogRequest();
        while (reader.pos < end) {
            var tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.log = exports.Log.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
    fromJSON: function (object) {
        return { log: isSet(object.log) ? exports.Log.fromJSON(object.log) : undefined };
    },
    toJSON: function (message) {
        var obj = {};
        message.log !== undefined &&
            (obj.log = message.log ? exports.Log.toJSON(message.log) : undefined);
        return obj;
    },
    fromPartial: function (object) {
        var message = createBaseLogRequest();
        message.log =
            object.log !== undefined && object.log !== null
                ? exports.Log.fromPartial(object.log)
                : undefined;
        return message;
    }
};
function createBaseLogResponse() {
    return { message: "" };
}
exports.LogResponse = {
    encode: function (message, writer) {
        if (writer === void 0) { writer = minimal_js_1["default"].Writer.create(); }
        if (message.message !== "") {
            writer.uint32(10).string(message.message);
        }
        return writer;
    },
    decode: function (input, length) {
        var reader = input instanceof minimal_js_1["default"].Reader ? input : new minimal_js_1["default"].Reader(input);
        var end = length === undefined ? reader.len : reader.pos + length;
        var message = createBaseLogResponse();
        while (reader.pos < end) {
            var tag = reader.uint32();
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
    fromJSON: function (object) {
        return { message: isSet(object.message) ? String(object.message) : "" };
    },
    toJSON: function (message) {
        var obj = {};
        message.message !== undefined && (obj.message = message.message);
        return obj;
    },
    fromPartial: function (object) {
        var _a;
        var message = createBaseLogResponse();
        message.message = (_a = object.message) !== null && _a !== void 0 ? _a : "";
        return message;
    }
};
exports.LogServiceService = {
    insertLog: {
        path: "/logpb.LogService/InsertLog",
        requestStream: false,
        responseStream: false,
        requestSerialize: function (value) {
            return Buffer.from(exports.LogRequest.encode(value).finish());
        },
        requestDeserialize: function (value) { return exports.LogRequest.decode(value); },
        responseSerialize: function (value) {
            return Buffer.from(exports.LogResponse.encode(value).finish());
        },
        responseDeserialize: function (value) { return exports.LogResponse.decode(value); }
    }
};
exports.LogServiceClient = (0, grpc_js_1.makeGenericClientConstructor)(exports.LogServiceService, "logpb.LogService");
function isSet(value) {
    return value !== null && value !== undefined;
}
