import{d as v,al as _,B as o,I as h,o as l,c,m as y,r,f as A,e as O,w as b,p as x,am as I,n as B,_ as p,an as $}from"./index-Eji0C-Q5.js";const C=["aria-expanded"],E={key:0,class:"accordion-item-content","data-testid":"accordion-item-content"},L=v({__name:"AccordionItem",setup(s){const e=_("parentAccordion"),t=o(null),n=h(()=>e===void 0?!1:e.multipleOpen&&Array.isArray(e.active.value)&&t.value!==null?e.active.value.includes(t.value):t.value===e.active.value);e!==void 0&&(t.value=e.count.value++);function i(){n.value?u():f()}function u(){e!==void 0&&(e.multipleOpen&&Array.isArray(e.active.value)&&t.value!==null?e.active.value.splice(e.active.value.indexOf(t.value),1):e.active.value=null)}function f(){e!==void 0&&(e.multipleOpen&&Array.isArray(e.active.value)&&t.value!==null?e.active.value.push(t.value):e.active.value=t.value)}function d(a){a instanceof HTMLElement&&(a.style.height=`${a.scrollHeight}px`)}function m(a){a instanceof HTMLElement&&(a.style.height="auto")}return(a,T)=>(l(),c("li",{class:B(["accordion-item",{active:n.value}])},[y("button",{class:"accordion-item-header",type:"button","aria-expanded":n.value?"true":"false","data-testid":"accordion-item-button",onClick:i},[r(a.$slots,"accordion-header",{},void 0,!0)],8,C),A(),O(I,{name:"accordion",onEnter:d,onAfterEnter:m,onBeforeLeave:d},{default:b(()=>[n.value?(l(),c("div",E,[r(a.$slots,"accordion-content",{},void 0,!0)])):x("",!0)]),_:3})],2))}}),g=p(L,[["__scopeId","data-v-53a0b6ce"]]),N={class:"accordion-list"},k=v({__name:"AccordionList",props:{initiallyOpen:{type:[Number,Array],required:!1,default:null},multipleOpen:{type:Boolean,required:!1,default:!1}},setup(s){const e=s,t=o(0),n=o(e.initiallyOpen!==null?e.initiallyOpen:e.multipleOpen?[]:null);return $("parentAccordion",{multipleOpen:e.multipleOpen,active:n,count:t}),(i,u)=>(l(),c("ul",N,[r(i.$slots,"default",{},void 0,!0)]))}}),w=p(k,[["__scopeId","data-v-bdbadd5e"]]);export{g as A,w as a};
