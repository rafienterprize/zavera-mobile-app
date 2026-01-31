'use client';

import { useState } from 'react';
import { Plus, X, Upload, Image as ImageIcon, Package } from 'lucide-react';

interface ProductVariant {
  id?: number;
  size: string;
  color: string;
  color_hex: string;
  stock: number;
  price: number;
  sku: string;
  weight: number;
  length: number;
  width: number;
  height: number;
  images: string[];
}

interface ProductFormData {
  name: string;
  description: string;
  category: string;
  subcategory: string;
  base_price: number;
  material: string;
  pattern: string;
  fit: string;
  sleeve: string;
  brand: string;
  images: string[];
  variants: ProductVariant[];
}

const CATEGORIES = {
  wanita: ['Dress', 'Blouse', 'Pants', 'Skirt', 'Jacket', 'Accessories'],
  pria: ['Shirt', 'T-Shirt', 'Pants', 'Jacket', 'Shoes', 'Accessories'],
  anak: ['Tops', 'Bottoms', 'Dress', 'Outerwear', 'Shoes'],
  sports: ['Activewear', 'Shoes', 'Accessories', 'Equipment'],
  luxury: ['Designer', 'Premium', 'Limited Edition'],
  beauty: ['Skincare', 'Makeup', 'Fragrance', 'Tools'],
};

const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL', 'XXXL'];
const COLORS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#000080' },
  { name: 'Red', hex: '#FF0000' },
  { name: 'Blue', hex: '#0000FF' },
  { name: 'Green', hex: '#008000' },
  { name: 'Yellow', hex: '#FFFF00' },
  { name: 'Pink', hex: '#FFC0CB' },
  { name: 'Gray', hex: '#808080' },
  { name: 'Brown', hex: '#A52A2A' },
];

export default function ProductFormComplete() {
  const [formData, setFormData] = useState<ProductFormData>({
    name: '',
    description: '',
    category: 'wanita',
    subcategory: '',
    base_price: 0,
    material: '',
    pattern: '',
    fit: '',
    sleeve: '',
    brand: '',
    images: [],
    variants: [],
  });

  const [currentStep, setCurrentStep] = useState(1);
  const [uploading, setUploading] = useState(false);

  return (
    <div className="max-w-7xl mx-auto p-6">
      {/* Progress Steps */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          {[
            { num: 1, label: 'Basic Info' },
            { num: 2, label: 'Images' },
            { num: 3, label: 'Variants' },
            { num: 4, label: 'Review' },
          ].map((step, idx) => (
            <div key={step.num} className="flex items-center flex-1">
              <div className="flex flex-col items-center flex-1">
                <div
                  className={`w-10 h-10 rounded-full flex items-center justify-center font-bold ${
                    currentStep >= step.num
                      ? 'bg-emerald-500 text-white'
                      : 'bg-gray-200 text-gray-500'
                  }`}
                >
                  {step.num}
                </div>
                <span className="text-sm mt-2">{step.label}</span>
              </div>
              {idx < 3 && (
                <div
                  className={`h-1 flex-1 ${
                    currentStep > step.num ? 'bg-emerald-500' : 'bg-gray-200'
                  }`}
                />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Step Content */}
      <div className="bg-white rounded-lg shadow-lg p-8">
        {currentStep === 1 && <BasicInfoStep formData={formData} setFormData={setFormData} />}
        {currentStep === 2 && <ImagesStep formData={formData} setFormData={setFormData} />}
        {currentStep === 3 && <VariantsStep formData={formData} setFormData={setFormData} />}
        {currentStep === 4 && <ReviewStep formData={formData} />}
      </div>

      {/* Navigation */}
      <div className="mt-6 flex justify-between">
        <button
          onClick={() => setCurrentStep(Math.max(1, currentStep - 1))}
          disabled={currentStep === 1}
          className="px-6 py-3 border rounded-lg disabled:opacity-50"
        >
          Previous
        </button>
        <button
          onClick={() => setCurrentStep(Math.min(4, currentStep + 1))}
          className="px-6 py-3 bg-emerald-500 text-white rounded-lg"
        >
          {currentStep === 4 ? 'Create Product' : 'Next'}
        </button>
      </div>
    </div>
  );
}

// Step Components will be in separate files
function BasicInfoStep({ formData, setFormData }: any) {
  return <div>Basic Info Step - See ProductFormBasicInfo.tsx</div>;
}

function ImagesStep({ formData, setFormData }: any) {
  return <div>Images Step - See ProductFormImages.tsx</div>;
}

function VariantsStep({ formData, setFormData }: any) {
  return <div>Variants Step - See ProductFormVariants.tsx</div>;
}

function ReviewStep({ formData }: any) {
  return <div>Review Step - See ProductFormReview.tsx</div>;
}
