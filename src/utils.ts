import fs from "fs";
export const loadFile = (path: string): string => {
  return fs.readFileSync(path, "utf8");
};
