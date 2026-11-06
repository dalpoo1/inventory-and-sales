export default function EmptyState({ message }: { message: string }) {
  return (
    <div className="rounded-md border border-dashed border-gray-300 bg-white p-8 text-center">
      <p className="text-sm text-gray-500">{message}</p>
    </div>
  )
}
