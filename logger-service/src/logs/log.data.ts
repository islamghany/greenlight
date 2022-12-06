import { model, Schema } from "mongoose";

export interface Log {
  service_name: string;
  error_message: string;
  stack_trace?: string | undefined;
}
const LogSchema = new Schema({
  service_name: {
    type: String,
    required: true,
  },
  stack_trace: {
    type: String,
  },
  error_message: {
    type: String,
    required: true,
  },
});

export default model("logs", LogSchema);
