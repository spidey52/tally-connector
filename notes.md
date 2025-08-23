Use of “ISDEEMEDPOSITIVE" method of Voucher object
The method “IsDeemedPositive” is used to identify whether the Amount should be Debited or Credited. This method has to be set to “Yes” if the amount should be Debited and “No” if the amount should be credited.

- sale vouchers

  - ledgerentries_list:

    - party
      - deemed_positive Yes
      - amount negative
    - discount
      - deemed_positive No
      - amount negative

  - all invetory list

    - products

      - amount positie
      - rate postive
      - bill positive
      - actual positive

    - accounting_allocation
      - amount postivie
      - deemed_positive no

- bank voucher

  - party

    - deemed_positive No
    - amount positive

  - bank
    - deemed_positive Yes
    - amount negative

check if bank allocation is created or not.

<!-- TODO: negative discount deke... deemed postive karenge to discount negative me lagega ya nii check karna hai. -->
<!-- TODO:   -->

one more view, like how much qty total he sold.
