'use client';

const CATEGORIES = {
  wanita: {
    label: 'Wanita',
    subcategories: ['Dress', 'Blouse', 'Pants', 'Skirt', 'Jacket', 'Accessories', 'Shoes', 'Bags'],
  },
  pria: {
    label: 'Pria',
    subcategories: ['Shirt', 'T-Shirt', 'Pants', 'Jeans', 'Jacket', 'Shoes', 'Accessories'],
  },
  anak: {
    label: 'Anak',
    subcategories: ['Tops', 'Bottoms', 'Dress', 'Outerwear', 'Shoes', 'Accessories'],
  },
  sports: {
    label: 'Sports',
    subcategories: ['Activewear', 'Running', 'Training', 'Shoes', 'Accessories', 'Equipment'],
  },
  luxury: {
    label: 'Luxury',
    subcategories: ['Designer', 'Premium', 'Limited Edition', 'Haute Couture'],
  },
  beauty: {
    label: 'Beauty',
    subcategories: ['Skincare', 'Makeup', 'Fragrance', 'Hair Care', 'Tools'],
  },
};

const MATERIALS = ['Cotton', 'Polyester', 'Wool', 'Silk', 'Linen', 'Denim', 'Leather', 'Synthetic', 'Mixed'];
const PATTERNS = ['Solid', 'Striped', 'Checked', 'Floral', 'Geometric', 'Abstract', 'Printed', 'Embroidered'];
const FITS = ['Slim', 'Regular', 'Relaxed', 'Oversized', 'Tailored', 'Loose'];
const SLEEVES = ['Sleeveless', 'Short Sleeve', '3/4 Sleeve', 'Long Sleeve', 'Cap Sleeve'];

interface Props {
  formData: any;
  setFormData: (data: any) => void;
}

export default function ProductFormBasicInfo({ formData, setFormData }: Props) {
  const handleChange = (field: string, value: any) => {
    setFormData({ ...formData, [field]: value });
  };

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Basic Information</h2>

      {/* Product Name */}
      <div>
        <label className="block text-sm font-medium mb-2">
          Product Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => handleChange('name', e.target.value)}
          placeholder="e.g., Classic Denim Jacket"
          className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
          required
        />
      </div>

      {/* Description */}
      <div>
        <label className="block text-sm font-medium mb-2">
          Description <span className="text-red-500">*</span>
        </label>
        <textarea
          value={formData.description}
          onChange={(e) => handleChange('description', e.target.value)}
          placeholder="Describe your product in detail..."
          rows={5}
          className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
          required
        />
        <p className="text-sm text-gray-500 mt-1">Min. 50 characters</p>
      </div>

      {/* Category & Subcategory */}
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium mb-2">
            Category <span className="text-red-500">*</span>
          </label>
          <select
            value={formData.category}
            onChange={(e) => {
              handleChange('category', e.target.value);
              handleChange('subcategory', ''); // Reset subcategory
            }}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
            required
          >
            {Object.entries(CATEGORIES).map(([key, cat]) => (
              <option key={key} value={key}>
                {cat.label}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">
            Subcategory <span className="text-red-500">*</span>
          </label>
          <select
            value={formData.subcategory}
            onChange={(e) => handleChange('subcategory', e.target.value)}
            className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
            required
          >
            <option value="">Select subcategory</option>
            {CATEGORIES[formData.category as keyof typeof CATEGORIES]?.subcategories.map((sub) => (
              <option key={sub} value={sub}>
                {sub}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Brand */}
      <div>
        <label className="block text-sm font-medium mb-2">Brand</label>
        <input
          type="text"
          value={formData.brand}
          onChange={(e) => handleChange('brand', e.target.value)}
          placeholder="e.g., Nike, Adidas, Zara"
          className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
        />
      </div>

      {/* Base Price */}
      <div>
        <label className="block text-sm font-medium mb-2">
          Base Price (IDR) <span className="text-red-500">*</span>
        </label>
        <input
          type="number"
          value={formData.base_price}
          onChange={(e) => handleChange('base_price', parseFloat(e.target.value))}
          placeholder="0"
          className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
          required
        />
        <p className="text-sm text-gray-500 mt-1">
          This is the base price. You can set different prices per variant later.
        </p>
      </div>

      {/* Product Attributes */}
      <div className="border-t pt-6">
        <h3 className="text-lg font-semibold mb-4">Product Attributes</h3>

        <div className="grid grid-cols-2 gap-4">
          {/* Material */}
          <div>
            <label className="block text-sm font-medium mb-2">Material</label>
            <select
              value={formData.material}
              onChange={(e) => handleChange('material', e.target.value)}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">Select material</option>
              {MATERIALS.map((mat) => (
                <option key={mat} value={mat}>
                  {mat}
                </option>
              ))}
            </select>
          </div>

          {/* Pattern */}
          <div>
            <label className="block text-sm font-medium mb-2">Pattern</label>
            <select
              value={formData.pattern}
              onChange={(e) => handleChange('pattern', e.target.value)}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">Select pattern</option>
              {PATTERNS.map((pat) => (
                <option key={pat} value={pat}>
                  {pat}
                </option>
              ))}
            </select>
          </div>

          {/* Fit */}
          <div>
            <label className="block text-sm font-medium mb-2">Fit</label>
            <select
              value={formData.fit}
              onChange={(e) => handleChange('fit', e.target.value)}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">Select fit</option>
              {FITS.map((fit) => (
                <option key={fit} value={fit}>
                  {fit}
                </option>
              ))}
            </select>
          </div>

          {/* Sleeve */}
          <div>
            <label className="block text-sm font-medium mb-2">Sleeve Length</label>
            <select
              value={formData.sleeve}
              onChange={(e) => handleChange('sleeve', e.target.value)}
              className="w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-emerald-500"
            >
              <option value="">Select sleeve</option>
              {SLEEVES.map((sleeve) => (
                <option key={sleeve} value={sleeve}>
                  {sleeve}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>
    </div>
  );
}
