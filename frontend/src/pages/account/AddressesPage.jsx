import React, { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import {
  MapPin,
  Plus,
  Edit2,
  Trash2,
  CheckCircle2,
  Home,
  Building,
  AlertCircle,
} from 'lucide-react';
import {
  FormField,
  Input,
  Button,
  Badge,
  Modal,
  SkeletonLoader,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import {
  addUserAddress,
  updateUserAddress,
  deleteUserAddress,
  setDefaultUserAddress,
} from '../../api/checkoutService';
import { validateName, validatePhone } from '../../utils/validators';
import './AddressesPage.css';

const INITIAL_FORM = {
  name: '',
  phone: '',
  address_line1: '',
  address_line_2: '',
  city: '',
  state: '',
  postal_code: '',
  country: 'India',
  is_default: false,
};

export function AddressesPage() {
  const context = useOutletContext() || {};
  const { addresses = [], isLoading = false, refreshData = () => {} } = context;
  const { showToast } = useToast();

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingAddressId, setEditingAddressId] = useState(null);
  const [formData, setFormData] = useState(INITIAL_FORM);
  const [formErrors, setFormErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Delete confirmation
  const [deleteCandidate, setDeleteCandidate] = useState(null);
  const [isDeleting, setIsDeleting] = useState(false);

  function openCreateModal() {
    setEditingAddressId(null);
    setFormData(INITIAL_FORM);
    setFormErrors({});
    setIsModalOpen(true);
  }

  function openEditModal(addr) {
    setEditingAddressId(addr.id);
    setFormData({
      name: addr.name || '',
      phone: addr.phone || '',
      address_line1: addr.address_line1 || '',
      address_line_2: addr.address_line_2 || addr.address_line2 || '',
      city: addr.city || '',
      state: addr.state || '',
      postal_code: addr.postal_code || '',
      country: addr.country || 'India',
      is_default: Boolean(addr.is_default),
    });
    setFormErrors({});
    setIsModalOpen(true);
  }

  function validateAddressForm() {
    const errors = {};
    const nameErr = validateName(formData.name);
    if (nameErr) errors.name = nameErr;

    const phoneErr = validatePhone(formData.phone);
    if (phoneErr) errors.phone = phoneErr;

    if (!formData.address_line1.trim()) {
      errors.address_line1 = 'Address line 1 is required';
    }
    if (!formData.city.trim()) {
      errors.city = 'City is required';
    }
    if (!formData.state.trim()) {
      errors.state = 'State is required';
    }
    if (!formData.postal_code.trim()) {
      errors.postal_code = 'Postal code is required';
    } else if (!/^\d{6}$/.test(formData.postal_code.trim())) {
      errors.postal_code = 'Enter a valid 6-digit PIN code';
    }

    return errors;
  }

  async function handleFormSubmit(e) {
    e.preventDefault();
    const errors = validateAddressForm();
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }
    setFormErrors({});
    setIsSubmitting(true);

    const payload = {
      name: formData.name.trim(),
      phone: formData.phone.trim(),
      address_line1: formData.address_line1.trim(),
      address_line_2: formData.address_line_2.trim(),
      city: formData.city.trim(),
      state: formData.state.trim(),
      postal_code: formData.postal_code.trim(),
      country: formData.country.trim() || 'India',
      is_default: formData.is_default,
    };

    try {
      if (editingAddressId) {
        await updateUserAddress(editingAddressId, payload);
        showToast({
          title: 'Address Updated',
          message: 'The delivery address was updated successfully.',
          type: 'success',
        });
      } else {
        await addUserAddress(payload);
        showToast({
          title: 'Address Added',
          message: 'New delivery address was added to your address book.',
          type: 'success',
        });
      }
      setIsModalOpen(false);
      await refreshData();
    } catch (err) {
      showToast({
        title: 'Save Failed',
        message: err.message || 'Unable to save address.',
        type: 'danger',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleSetDefault(id) {
    try {
      await setDefaultUserAddress(id);
      await refreshData();
      showToast({
        title: 'Default Set',
        message: 'Default delivery address updated.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Action Failed',
        message: err.message || 'Failed to set default address.',
        type: 'danger',
      });
    }
  }

  async function handleConfirmDelete() {
    if (!deleteCandidate) return;
    setIsDeleting(true);
    try {
      await deleteUserAddress(deleteCandidate.id);
      setDeleteCandidate(null);
      await refreshData();
      showToast({
        title: 'Address Deleted',
        message: 'The address has been removed from your book.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Delete Failed',
        message: err.message || 'Failed to remove address.',
        type: 'danger',
      });
    } finally {
      setIsDeleting(false);
    }
  }

  if (isLoading) {
    return (
      <div className="account-addresses-loading" data-testid="addresses-loading">
        <SkeletonLoader lines={6} />
      </div>
    );
  }

  return (
    <div className="account-addresses-page" data-testid="account-addresses-page">
      {/* Top Header Row */}
      <div className="account-addresses-header">
        <div>
          <h1 className="account-addresses-title">Address Book</h1>
          <p className="account-addresses-subtitle">
            Manage your saved delivery destinations for seamless tailoring delivery.
          </p>
        </div>
        <Button
          variant="primary"
          size="sm"
          leftIcon={<Plus size={16} />}
          onClick={openCreateModal}
          data-testid="add-new-address-btn"
        >
          ADD NEW ADDRESS
        </Button>
      </div>

      {/* Address Cards Grid matching Figma 32:11120 */}
      {addresses.length > 0 ? (
        <div className="addresses-grid" data-testid="addresses-grid">
          {addresses.map((addr) => {
            const isDefault = Boolean(addr.is_default);
            return (
              <div
                key={addr.id}
                className={`address-card ${isDefault ? 'is-default' : ''}`}
                data-testid={`address-card-${addr.id}`}
              >
                <div className="address-card__header">
                  <div className="address-card__title-group">
                    <MapPin size={16} className="address-card__icon" />
                    <strong className="address-card__recipient">{addr.name}</strong>
                  </div>
                  {isDefault && (
                    <Badge variant="neutral" size="sm">
                      DEFAULT
                    </Badge>
                  )}
                </div>

                <div className="address-card__body">
                  <p className="address-card__lines">
                    {addr.address_line1}
                    {(addr.address_line_2 || addr.address_line2) && (
                      <>, {addr.address_line_2 || addr.address_line2}</>
                    )}
                    <br />
                    {addr.city}, {addr.state} {addr.postal_code}
                    <br />
                    {addr.country || 'India'}
                  </p>
                  <p className="address-card__phone">
                    <span className="address-card__phone-label">Tel:</span> {addr.phone}
                  </p>
                </div>

                <div className="address-card__footer">
                  <div className="address-card__actions">
                    <Button
                      variant="ghost"
                      size="sm"
                      leftIcon={<Edit2 size={13} />}
                      onClick={() => openEditModal(addr)}
                      data-testid={`edit-address-${addr.id}`}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      leftIcon={<Trash2 size={13} />}
                      onClick={() => setDeleteCandidate(addr)}
                      data-testid={`delete-address-${addr.id}`}
                    >
                      Delete
                    </Button>
                  </div>

                  {!isDefault && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleSetDefault(addr.id)}
                      data-testid={`set-default-${addr.id}`}
                    >
                      Set As Default
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        <div className="account-empty-block" data-testid="addresses-empty-state">
          <MapPin size={36} className="account-empty-icon" />
          <h2 className="account-empty-title">Your Address Book is Empty</h2>
          <p className="account-empty-text">
            Add your primary shipping destination for expedited checkout and doorstep delivery.
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={openCreateModal}
            leftIcon={<Plus size={16} />}
          >
            ADD FIRST ADDRESS
          </Button>
        </div>
      )}

      {/* Add / Edit Address Modal */}
      <Modal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title={editingAddressId ? 'Edit Delivery Address' : 'Add New Address'}
        eyebrow="ADDRESS BOOK"
        size="md"
      >
        <form onSubmit={handleFormSubmit} className="address-modal-form" noValidate>
          <div className="address-form-grid">
            <FormField label="RECIPIENT NAME" htmlFor="addr-name" required error={formErrors.name}>
              <Input
                id="addr-name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Full Name"
                hasError={Boolean(formErrors.name)}
                data-testid="addr-input-name"
              />
            </FormField>

            <FormField label="CONTACT PHONE" htmlFor="addr-phone" required error={formErrors.phone}>
              <Input
                id="addr-phone"
                type="tel"
                value={formData.phone}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                placeholder="10-digit mobile number"
                hasError={Boolean(formErrors.phone)}
                data-testid="addr-input-phone"
              />
            </FormField>
          </div>

          <FormField
            label="ADDRESS LINE 1 (FLAT / STREET / BUILDING)"
            htmlFor="addr-line1"
            required
            error={formErrors.address_line1}
          >
            <Input
              id="addr-line1"
              value={formData.address_line1}
              onChange={(e) => setFormData({ ...formData, address_line1: e.target.value })}
              placeholder="e.g. 14 Mercer Court, Suite 4B"
              hasError={Boolean(formErrors.address_line1)}
              data-testid="addr-input-line1"
            />
          </FormField>

          <FormField
            label="ADDRESS LINE 2 (LANDMARK / LOCALITY — OPTIONAL)"
            htmlFor="addr-line2"
          >
            <Input
              id="addr-line2"
              value={formData.address_line_2}
              onChange={(e) => setFormData({ ...formData, address_line_2: e.target.value })}
              placeholder="e.g. Near Fort Heritage"
              data-testid="addr-input-line2"
            />
          </FormField>

          <div className="address-form-grid-3">
            <FormField label="CITY" htmlFor="addr-city" required error={formErrors.city}>
              <Input
                id="addr-city"
                value={formData.city}
                onChange={(e) => setFormData({ ...formData, city: e.target.value })}
                placeholder="City"
                hasError={Boolean(formErrors.city)}
                data-testid="addr-input-city"
              />
            </FormField>

            <FormField label="STATE" htmlFor="addr-state" required error={formErrors.state}>
              <Input
                id="addr-state"
                value={formData.state}
                onChange={(e) => setFormData({ ...formData, state: e.target.value })}
                placeholder="State"
                hasError={Boolean(formErrors.state)}
                data-testid="addr-input-state"
              />
            </FormField>

            <FormField
              label="POSTAL PIN CODE"
              htmlFor="addr-pincode"
              required
              error={formErrors.postal_code}
            >
              <Input
                id="addr-pincode"
                maxLength={6}
                value={formData.postal_code}
                onChange={(e) =>
                  setFormData({ ...formData, postal_code: e.target.value.replace(/\D/g, '') })
                }
                placeholder="600001"
                hasError={Boolean(formErrors.postal_code)}
                data-testid="addr-input-pincode"
              />
            </FormField>
          </div>

          <div className="address-default-toggle">
            <label className="address-default-checkbox-label">
              <input
                type="checkbox"
                checked={formData.is_default}
                onChange={(e) => setFormData({ ...formData, is_default: e.target.checked })}
                data-testid="addr-checkbox-default"
              />
              <span>Set as default delivery address</span>
            </label>
          </div>

          <div className="address-modal-actions">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={() => setIsModalOpen(false)}
            >
              CANCEL
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="md"
              isLoading={isSubmitting}
              data-testid="save-address-btn"
            >
              {editingAddressId ? 'SAVE CHANGES' : 'ADD ADDRESS'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={Boolean(deleteCandidate)}
        onClose={() => setDeleteCandidate(null)}
        title="Delete Address"
        eyebrow="CONFIRMATION"
        size="sm"
      >
        <div className="address-delete-modal" data-testid="delete-address-modal">
          <p className="address-delete-desc">
            Are you sure you wish to remove this delivery address? This action cannot be undone.
          </p>
          {deleteCandidate && (
            <div className="address-delete-preview">
              <strong>{deleteCandidate.name}</strong>
              <span>
                {deleteCandidate.address_line1}, {deleteCandidate.city}
              </span>
            </div>
          )}
          <div className="address-modal-actions">
            <Button
              variant="outline"
              size="md"
              onClick={() => setDeleteCandidate(null)}
              disabled={isDeleting}
            >
              KEEP ADDRESS
            </Button>
            <Button
              variant="primary"
              size="md"
              onClick={handleConfirmDelete}
              isLoading={isDeleting}
              data-testid="confirm-delete-address-btn"
            >
              DELETE
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
