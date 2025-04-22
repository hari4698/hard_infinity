import { Button } from "@/components/ui/button"; // Assuming this path is correct
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"; // Assuming this path is correct based on Shadcn setup
import { Checkbox } from "@/components/ui/checkbox"; // Assuming this path is correct
import { Input } from "@/components/ui/input"; // Assuming this path is correct
import { Label } from "@/components/ui/label"; // Assuming this path is correct
import { createFileRoute } from '@tanstack/react-router';
import { useState } from 'react';

interface Task {
  id: string;
  name: string;
  completed: boolean;
}

interface Section {
  id: string;
  name: string;
  tasks: Task[];
}

interface Challenge {
  id: string;
  name: string;
  sections: Section[];
}

export const Route = createFileRoute('/settings')({
  component: Settings,
});

function Settings() {
  const [challenges, setChallenges] = useState<Challenge[]>([]);
  const [selectedChallengeId, setSelectedChallengeId] = useState<string | null>(null);

  // --- State for Add Challenge Dialog ---
  const [newChallengeNameInput, setNewChallengeNameInput] = useState('');
  const [isAddChallengeDialogOpen, setIsAddChallengeDialogOpen] = useState(false);

  // --- State for Add Section Dialog ---
  const [newSectionNameInput, setNewSectionNameInput] = useState('');
  const [isAddSectionDialogOpen, setIsAddSectionDialogOpen] = useState(false);

  // --- State for Add Task Dialog ---
  const [newTaskNameInput, setNewTaskNameInput] = useState('');
  const [isAddTaskDialogOpen, setIsAddTaskDialogOpen] = useState(false);
  const [sectionTargetForTask, setSectionTargetForTask] = useState<{ challengeId: string; sectionId: string } | null>(null);


  // --- Add Challenge Logic ---
  const handleAddChallengeSubmit = () => {
    if (newChallengeNameInput.trim()) {
      const newChallenge: Challenge = {
        id: crypto.randomUUID(),
        name: newChallengeNameInput.trim(),
        sections: [],
      };
      setChallenges([...challenges, newChallenge]);
      setSelectedChallengeId(newChallenge.id); // Select the new challenge
      setNewChallengeNameInput(''); // Reset input
      setIsAddChallengeDialogOpen(false); // Close dialog
    }
  };

  const handleChallengeDialogOpenChange = (open: boolean) => {
    setIsAddChallengeDialogOpen(open);
    if (!open) {
        setNewChallengeNameInput(''); // Reset on close
    }
  };


  // --- Add Section Logic ---
  const handleAddSectionSubmit = (challengeId: string) => {
    if (newSectionNameInput.trim() && challengeId) {
      setChallenges(prevChallenges =>
        prevChallenges.map(challenge =>
          challenge.id === challengeId
            ? {
                ...challenge,
                sections: [
                  ...challenge.sections,
                  { id: crypto.randomUUID(), name: newSectionNameInput.trim(), tasks: [] },
                ],
              }
            : challenge
        )
      );
      setNewSectionNameInput(''); // Reset
      setIsAddSectionDialogOpen(false); // Close
    }
  };

  const handleSectionDialogOpenChange = (open: boolean) => {
    setIsAddSectionDialogOpen(open);
    if (!open) {
        setNewSectionNameInput(''); // Reset on close
    }
  };


  // --- Add Task Logic ---
  const prepareAddTaskDialog = (challengeId: string, sectionId: string) => {
    setSectionTargetForTask({ challengeId, sectionId });
    setIsAddTaskDialogOpen(true); // Open the dialog
  };

  const handleAddTaskSubmit = () => {
    if (newTaskNameInput.trim() && sectionTargetForTask) {
      const { challengeId, sectionId } = sectionTargetForTask;
      setChallenges(prevChallenges =>
          prevChallenges.map(challenge =>
              challenge.id === challengeId
                  ? {
                      ...challenge,
                      sections: challenge.sections.map(section =>
                          section.id === sectionId
                              ? {
                                  ...section,
                                  tasks: [
                                      ...section.tasks,
                                      { id: crypto.randomUUID(), name: newTaskNameInput.trim(), completed: false },
                                  ],
                                }
                              : section
                      ),
                    }
                  : challenge
          )
      );
      setNewTaskNameInput('');
      setIsAddTaskDialogOpen(false); // Close the dialog
      setSectionTargetForTask(null); // Reset target
    }
  };

  const handleTaskDialogOpenChange = (open: boolean) => {
    setIsAddTaskDialogOpen(open);
    if (!open) {
      setNewTaskNameInput('');
      setSectionTargetForTask(null);
    }
  };

  const handleChallengeNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewChallengeNameInput(e.currentTarget.value);
  };

  const handleSectionNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewSectionNameInput(e.currentTarget.value);
  };

  const handleTaskNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setNewTaskNameInput(e.currentTarget.value);
  };

  // --- Task Completion Logic ---
  const handleTaskCompletionChange = (taskId: string, completed: boolean | "indeterminate") => { // Accept boolean or indeterminate
    setChallenges(prevChallenges =>
      prevChallenges.map(challenge => ({
        ...challenge,
        sections: challenge.sections.map(section => ({
          ...section,
          tasks: section.tasks.map(task =>
            task.id === taskId ? { ...task, completed: !!completed } : task // Force to boolean
          ),
        })),
      }))
    );
  };

  // --- Component Rendering ---
  const selectedChallenge = challenges.find(c => c.id === selectedChallengeId);

  return (
    <div>
      {/* --- Challenges Row + Add Button/Dialog --- */}
      <div className="mb-4 flex flex-wrap gap-4 items-center"> {/* Use items-center for better alignment */}
        {challenges.map(challenge => (
          <Button // Use Shadcn Button for consistency
            key={challenge.id}
            variant={selectedChallengeId === challenge.id ? 'default' : 'secondary'} // Example variants
            size="lg" // Match size roughly
            className={`aspect-square font-bold rounded-lg ${selectedChallengeId === challenge.id ? 'ring-2 ring-ring ring-offset-2' : ''}`} // Use Shadcn ring utils
            onClick={() => setSelectedChallengeId(challenge.id)}
          >
            {challenge.name}
          </Button>
        ))}

        <Dialog open={isAddChallengeDialogOpen} onOpenChange={handleChallengeDialogOpenChange}>
          <DialogTrigger asChild>
             <Button variant="outline" size="lg" className="aspect-square font-bold rounded-lg border-dashed border-2 border-green-500 text-green-500 hover:bg-green-50 hover:text-green-600"> {/* Style as add button */}
               + {/* Simple plus icon */}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add New Challenge</DialogTitle>
              <DialogDescription>
                Enter a name for your new challenge. Press Enter or click Add.
              </DialogDescription>
            </DialogHeader>
            {/* Use form for semantic correctness and Enter submission */}
            <form onSubmit={(e) => { e.preventDefault(); handleAddChallengeSubmit(); }}>
                <div className="grid gap-4 py-4">
                    <div className="grid grid-cols-4 items-center gap-4">
                    <Label htmlFor="challenge-name" className="text-right">
                        Name
                    </Label>
                    <Input
                        id="challenge-name"
                        value={newChallengeNameInput}
                        onChange={handleChallengeNameChange}
                        className="col-span-3"
                        autoFocus // Focus input when dialog opens
                    />
                    </div>
                </div>
                <DialogFooter>
                  <DialogClose asChild>
                    <Button type="button" variant="outline">Cancel</Button>
                  </DialogClose>
                  <Button type="submit">Add Challenge</Button>
                </DialogFooter>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {/* --- Selected Challenge Details --- */}
      {selectedChallenge && (
        <div className="mt-6 p-4 border rounded-lg bg-card text-card-foreground shadow"> {/* Use theme vars */}
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">{selectedChallenge.name}</h2>
            {/* --- Add Section Button/Dialog --- */}
            <Dialog open={isAddSectionDialogOpen} onOpenChange={handleSectionDialogOpenChange}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm"> {/* Use Shadcn Button */}
                  + Add Section
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Add New Section</DialogTitle>
                  <DialogDescription>
                    Enter a name for the new section (e.g., Morning Routine). Press Enter or click Add.
                  </DialogDescription>
                </DialogHeader>
                 <form onSubmit={(e) => { e.preventDefault(); handleAddSectionSubmit(selectedChallenge.id); }}>
                    <div className="grid gap-4 py-4">
                        <div className="grid grid-cols-4 items-center gap-4">
                        <Label htmlFor="section-name" className="text-right">
                            Name
                        </Label>
                        <Input
                            id="section-name"
                            value={newSectionNameInput}
                            onChange={handleSectionNameChange}
                            className="col-span-3"
                            autoFocus
                        />
                        </div>
                    </div>
                    <DialogFooter>
                      <DialogClose asChild>
                          <Button type="button" variant="outline">Cancel</Button>
                      </DialogClose>
                      <Button type="submit">Add Section</Button>
                    </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          </div>


          {/* --- Sections and Tasks --- */}
          {/* Place Task Dialog *outside* the map to avoid multiple instances */}
          <Dialog open={isAddTaskDialogOpen} onOpenChange={handleTaskDialogOpenChange}>
            <DialogContent>
              <DialogHeader>
                  <DialogTitle>Add New Task</DialogTitle>
                  <DialogDescription>
                      Enter a name for the new task. Press Enter or click Add.
                  </DialogDescription>
                </DialogHeader>
                <form onSubmit={(e) => { e.preventDefault(); handleAddTaskSubmit(); }}>
                  <div className="grid gap-4 py-4">
                      <div className="grid grid-cols-4 items-center gap-4">
                      <Label htmlFor="task-name" className="text-right">
                          Name
                      </Label>
                      <Input
                          id="task-name"
                          value={newTaskNameInput}
                          onChange={handleTaskNameChange}
                          className="col-span-3"
                          autoFocus
                      />
                      </div>
                  </div>
                  <DialogFooter>
                    <DialogClose asChild>
                        <Button type="button" variant="outline">Cancel</Button>
                    </DialogClose>
                    <Button type="submit">Add Task</Button>
                  </DialogFooter>
                </form>
            </DialogContent>
          </Dialog>

          <div className="space-y-4"> {/* Add spacing between sections */}
            {selectedChallenge.sections.map(section => (
              <div key={section.id} className="p-3 border rounded bg-muted/40"> {/* Use theme vars */}
                <div className="flex justify-between items-center mb-2">
                  <h3 className="text-lg font-medium">{section.name}</h3>
                  {/* --- Add Task Button (triggers the single dialog) --- */}
                  <Button
                      variant="ghost" // Subtle button
                      size="sm"
                      onClick={() => prepareAddTaskDialog(selectedChallenge.id, section.id)} // Prepare the dialog context
                      className="text-xs" // Smaller text
                  >
                      + Add Task
                  </Button>
                </div>
                <div className="pl-2 space-y-2 text-sm"> {/* Adjust spacing */}
                  {section.tasks.length === 0 && <p className="text-muted-foreground italic">No tasks yet.</p>}
                  {section.tasks.map(task => (
                      <div key={task.id} className="flex items-center space-x-2">
                          <Checkbox 
                              id={`task-${task.id}`}
                              checked={task.completed}
                              onCheckedChange={(checked) => handleTaskCompletionChange(task.id, checked)} // Pass checked state directly
                          />
                          <Label 
                              htmlFor={`task-${task.id}`} 
                              className={`flex-grow ${task.completed ? 'line-through text-muted-foreground' : ''}`} // Style completed tasks
                          >
                              {task.name}
                          </Label>
                      </div>
                  ))}
                </div>
              </div>
            ))}
             {selectedChallenge.sections.length === 0 && <p className="text-muted-foreground italic">No sections added yet.</p>}
          </div>
        </div>
      )}
    </div>
  );
}
