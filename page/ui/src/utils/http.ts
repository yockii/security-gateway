import axios, {AxiosInstance} from "axios";
// import { useUserStore } from "@/store/modules/user";

const instance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_APP_API_BASE_URL,
  timeout: 5000,
  //   withCredentials: true,
  headers: {
    "Content-Type": "application/json",
  },
});

// let tokenState = 0; // 0: 未获取token，1: 正在获取token，2: 已获取token
// let requestQueue: any[] = [];

instance.interceptors.request.use(
    async (config) => {
      return config;
      // 除了登录接口，其他接口都需要携带token
      // if (tokenState === 2 || config.url === "/zzd/sso/ssoLogin") {
      //   const token = useUserStore().token;
      //   token && (config.headers.Authorization = `Bearer ${token}`);
      //   return config;
      // } else {
      //   if (tokenState === 0) {
      //     tokenState = 1;
      //     try {
      //       const token = await useUserStore().getToken();
      //       token && (config.headers.Authorization = `Bearer ${token}`);
      //       tokenState = 2;
      //       requestQueue.forEach((cb) => cb());
      //       requestQueue = [];
      //     } catch (error) {
      //       console.error(error);
      //     } finally {
      //       if (tokenState !== 2) {
      //         tokenState = 0;
      //       }
      //       return config;
      //     }
      //   } else {
      //     return new Promise((resolve) => {
      //       requestQueue.push(() => {
      //         const token = useUserStore().token;
      //         token && (config.headers.Authorization = `Bearer ${token}`);
      //         resolve(config);
      //       });
      //     });
      //   }
      // }
    },
    (error: any) => {
      return Promise.reject(error);
    }
);

instance.interceptors.response.use(
    (response) => {
      return response;
    },
    (error) => {
      return Promise.reject(error);
    }
);

export default instance;
