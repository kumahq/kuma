import{e as _,L as d,o,f as s,c as a,a as m,w as p,g as f,G as y,r as e}from"./index.f4381a04.js";import{_ as B}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.a081fa47.js";import{E as k}from"./ErrorBlock.3c391f50.js";import{_ as u}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.52c551fa.js";const L={name:"LabelList",components:{EmptyBlock:B,ErrorBlock:k,LoadingBlock:u,KCard:d},props:{items:{type:Object,default:null},title:{type:String,default:null},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}}},E={key:3,class:"label-list-content"},b={class:"label-list__col-wrapper multi-col"};function g(n,v,t,h,C,x){const l=e("LoadingBlock"),r=e("ErrorBlock"),c=e("EmptyBlock"),i=e("KCard");return o(),s("div",null,[t.isLoading?(o(),a(l,{key:0})):t.hasError?(o(),a(r,{key:1})):t.isEmpty?(o(),a(c,{key:2})):(o(),s("div",E,[m(i,{"border-variant":"noBorder"},{body:p(()=>[f("div",b,[y(n.$slots,"default")])]),_:3})]))])}const S=_(L,[["render",g]]);export{S as L};
