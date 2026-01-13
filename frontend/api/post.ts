import { axiosInstance } from "@/api/instance";

type PathsAccepted = string;

export const post = async <T, D = unknown>(
  path: PathsAccepted,
  data?: D
): Promise<T> => {
  const response = await axiosInstance.post<T>(path, data);
  return response.data;
};
