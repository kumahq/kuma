import{K as u}from"./index-9dd3e7d3.js";import{d as p,o as f,b as y,w as n,e as r,q as i,ae as _,f as x,r as T,af as h,P as g,ag as C,p as m,t as B,_ as b}from"./index-57969804.js";const S={class:"visually-hidden"},v={inheritAttrs:!1},q=p({...v,__name:"CopyButton",props:{text:{type:String,required:!1,default:""},getText:{type:Function,required:!1,default:null},copyText:{type:String,required:!1,default:"Copy"},tooltipSuccessText:{type:String,required:!1,default:"Copied code!"},tooltipFailText:{type:String,required:!1,default:"Failed to copy!"},hasBorder:{type:Boolean,default:!1},hideTitle:{type:Boolean,default:!1}},setup(d){const t=d;async function c(a,l){const e=a.currentTarget;let o=!1;try{const s=t.getText?await t.getText():t.text;o=await l(s)}catch{o=!1}finally{const s=o?t.tooltipSuccessText:t.tooltipFailText;e instanceof HTMLButtonElement&&(e.setAttribute("data-tooltip-copy-success",String(o)),e.setAttribute("data-tooltip-text",s),window.setTimeout(function(){e instanceof HTMLButtonElement&&e.removeAttribute("data-tooltip-text")},1500))}}return(a,l)=>(f(),y(i(C),null,{default:n(({copyToClipboard:e})=>[r(i(g),h(a.$attrs,{appearance:"tertiary",class:["copy-button",{"non-visual-button":!t.hasBorder}],"data-testid":"copy-button",title:t.hideTitle?void 0:t.copyText,type:"button",onClick:o=>c(o,e)}),{default:n(()=>[r(i(_),{size:i(u),title:t.hideTitle?void 0:t.copyText,"hide-title":t.hideTitle},null,8,["size","title","hide-title"]),x(),T(a.$slots,"default",{},()=>[m("span",S,B(t.copyText),1)],!0)]),_:2},1040,["class","title","onClick"])]),_:3}))}});const A=b(q,[["__scopeId","data-v-bd6d6132"]]);export{A as C};
