import { Checkbox } from "@/components/ui/checkbox"; // Assuming path
import { Input } from "@/components/ui/input";     // Assuming path
import { Label } from "@/components/ui/label";     // Assuming path
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"; // Assuming path
import { Textarea } from "@/components/ui/textarea"; // Assuming path

export default function Tasks() {
  // Note: State management (e.g., using useState) is not implemented here.
  // This component currently only renders the UI structure.
  return (
    <div className="space-y-8 p-4 md:p-6"> {/* Add padding and spacing */}
      <h1 className="text-2xl font-bold">Daily Tasks</h1> {/* Main heading */}

      <section>
        <h2 className="text-xl font-semibold mb-4 border-b pb-2">Morning Routine</h2>
        <ul className="space-y-4"> {/* Spacing between tasks */}
          <li className="flex items-center space-x-3"> {/* Increased spacing */}
            <Checkbox id="morning-photo" />
            <Label htmlFor="morning-photo" className="font-normal">Take progress photo</Label>
          </li>
          <li className="flex items-center space-x-3">
            <Checkbox id="morning-meditation" />
            <Label htmlFor="morning-meditation" className="font-normal">Complete 20-minute meditation</Label>
          </li>
          <li className="flex items-center space-x-3">
            <Checkbox id="morning-social" />
            <Label htmlFor="morning-social" className="font-normal">Avoid social media before 11 AM</Label>
          </li>
          <li className="flex items-center space-x-3">
            <Checkbox id="morning-meals" />
            <Label htmlFor="morning-meals" className="font-normal">Plan meals for tomorrow</Label>
          </li>
        </ul>
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-4 border-b pb-2">Physical Activity</h2>
        <ul className="space-y-6"> {/* Increased spacing for complex items */}
          <li className="flex items-center space-x-3">
            <Checkbox id="activity-complete" />
            <Label htmlFor="activity-complete" className="font-normal">Complete workout/active rest</Label>
          </li>
          <li className="grid w-full max-w-sm items-center gap-1.5"> {/* Grid layout for label+input */}
            <Label htmlFor="activity-type">Activity type</Label>
            <Select>
              <SelectTrigger id="activity-type" className="w-[180px]">
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="strength">Strength</SelectItem>
                <SelectItem value="cardio">Cardio</SelectItem>
                <SelectItem value="rest">Active Rest</SelectItem>
              </SelectContent>
            </Select>
          </li>
           <li className="grid w-full max-w-sm items-center gap-1.5">
            <Label htmlFor="activity-duration">Duration (minutes)</Label>
            <Input type="number" id="activity-duration" placeholder="e.g., 45" className="w-[180px]" />
          </li>
          <li className="grid w-full max-w-sm items-center gap-1.5">
            <Label htmlFor="activity-intensity">Intensity (1-10)</Label>
            <Input type="number" id="activity-intensity" placeholder="e.g., 7" min="1" max="10" className="w-[180px]" />
          </li>
          <li className="grid w-full items-center gap-1.5">
             <Label htmlFor="activity-notes">Notes</Label>
             <Textarea id="activity-notes" placeholder="Any details..." />
          </li>
        </ul>
      </section>
    </div>
  );
}
