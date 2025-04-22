// app/routes/index.tsx
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { Separator } from "@/components/ui/separator";
import Tasks from "@/components/Tasks";
import DateNavigator from "@/components/DateNavigator";

export const Route = createFileRoute("/")({
  component: Home,
});

function Home() {
  const router = useRouter();
  const state = Route.useLoaderData();

  return (
    <div className="flex flex-col">
      <DateNavigator />
      <Separator  />
      <div className="m-4 p-4  flex flex-col">
        <Tasks />
      </div>
    </div>
  );
}
