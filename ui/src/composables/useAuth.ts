import { ref } from "vue";
import { pb } from "@/lib/pocketbase";
import type { UsersRecord } from "@/generated/pocketbase-types";

const user = ref(pb.authStore.record as UsersRecord | null);
const isAuthenticated = ref(pb.authStore.isValid);

pb.authStore.onChange((token, record) => {
  user.value = record as UsersRecord | null;
  console.log("user.value changed")
  isAuthenticated.value = !!token;
  console.log("CHANGEED");
});

export function useAuth() {
  const login = () =>
    pb
      .collection("users")
      .authWithPassword("mati852456@gmail.com", "bFKL1rYFgaifPrR1KmSj");

  const logout = () => {
    pb.authStore.clear();
  };

  return {
    user,
    isAuthenticated,
    login,
    logout,
  };
}
