import { useState, useRef } from 'react'
import { Upload, X, Image as ImageIcon } from 'lucide-react'

interface PhotoSelectorProps {
  onFilesSelected: (files: File[]) => void
  maxFiles?: number
  maxSizeMB?: number
  disabled?: boolean
}

export default function PhotoSelector({ 
  onFilesSelected, 
  maxFiles = 5, 
  maxSizeMB = 5, 
  disabled = false
}: PhotoSelectorProps) {
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [previews, setPreviews] = useState<string[]>([])
  const [error, setError] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    setError('')

    // Validate file count
    if (selectedFiles.length + files.length > maxFiles) {
      setError(`Максимум ${maxFiles} фото`)
      return
    }

    // Validate file types and sizes
    const validFiles: File[] = []
    const newPreviews: string[] = []

    files.forEach((file) => {
      // Check file type
      if (!file.type.startsWith('image/')) {
        setError(`${file.name} не є зображенням`)
        return
      }

      // Check file size
      if (file.size > maxSizeMB * 1024 * 1024) {
        setError(`${file.name} занадто великий (макс ${maxSizeMB}МБ)`)
        return
      }

      validFiles.push(file)
      const reader = new FileReader()
      reader.onload = (e) => {
        if (e.target?.result) {
          newPreviews.push(e.target.result as string)
          setPreviews((prev) => [...prev, ...newPreviews])
        }
      }
      reader.readAsDataURL(file)
    })

    const updatedFiles = [...selectedFiles, ...validFiles]
    setSelectedFiles(updatedFiles)
    onFilesSelected(updatedFiles)
    e.target.value = '' // Reset input
  }

  const removeFile = (index: number) => {
    const updatedFiles = selectedFiles.filter((_, i) => i !== index)
    const updatedPreviews = previews.filter((_, i) => i !== index)
    setSelectedFiles(updatedFiles)
    setPreviews(updatedPreviews)
    onFilesSelected(updatedFiles)
  }

  return (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Фотографії <span className="text-gray-500 text-xs">(макс {maxFiles} фото, {maxSizeMB}МБ кожне)</span>
        </label>

        {/* File input */}
        <div className="flex items-center space-x-4">
          <input
            ref={fileInputRef}
            type="file"
            multiple
            accept="image/*"
            onChange={handleFileSelect}
            disabled={disabled || selectedFiles.length >= maxFiles}
            className="hidden"
            id="photo-selector"
          />
          <label
            htmlFor="photo-selector"
            className={`flex items-center space-x-2 px-4 py-2 border border-gray-300 rounded-md cursor-pointer hover:bg-gray-50 ${
              disabled || selectedFiles.length >= maxFiles
                ? 'opacity-50 cursor-not-allowed'
                : ''
            }`}
          >
            <Upload className="h-4 w-4" />
            <span>Вибрати фото</span>
          </label>
          {selectedFiles.length > 0 && (
            <span className="text-sm text-gray-600">
              Вибрано: {selectedFiles.length} / {maxFiles}
            </span>
          )}
        </div>

        {error && (
          <p className="mt-2 text-sm text-red-600">{error}</p>
        )}

        {/* Preview grid */}
        {previews.length > 0 && (
          <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
            {previews.map((preview, index) => (
              <div key={index} className="relative group">
                <img
                  src={preview}
                  alt={`Preview ${index + 1}`}
                  className="w-full h-32 object-cover rounded-lg border border-gray-300"
                />
                <button
                  type="button"
                  onClick={() => removeFile(index)}
                  className="absolute top-2 right-2 bg-red-500 text-white rounded-full p-1 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <X className="h-4 w-4" />
                </button>
                <div className="absolute bottom-2 left-2 bg-black/50 text-white text-xs px-2 py-1 rounded">
                  {(selectedFiles[index].size / 1024 / 1024).toFixed(2)} МБ
                </div>
              </div>
            ))}
          </div>
        )}

        {selectedFiles.length === 0 && (
          <div className="mt-4 border-2 border-dashed border-gray-300 rounded-lg p-8 text-center">
            <ImageIcon className="h-12 w-12 text-gray-400 mx-auto mb-2" />
            <p className="text-sm text-gray-500">Натисніть "Вибрати фото" для додавання</p>
          </div>
        )}
      </div>
    </div>
  )
}

