import LogModel, { Log } from "./log.data.js";

export async function InsertLog(log: Log): Promise<boolean> {
  const newLog = new LogModel(log);

  try {
    await newLog.save();
    return true;
  } catch (err) {
    throw err;
  }
}
