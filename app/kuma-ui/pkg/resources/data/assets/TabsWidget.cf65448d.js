import{d as h,r as y,e as v,o as s,i as n,c,F as B,K as u,A as T,b as f,j as d,a as p,L as k,n as E,u as l,w as m,H as S,I as w,m as q,Y as V,D as L,E as A,G as C}from"./index.0c4c6d47.js";import{E as I}from"./ErrorBlock.39782751.js";import{_ as N}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.e84843f4.js";const $=e=>(L("data-v-36f2b969"),e=e(),A(),e),x={class:"tab-container","data-testid":"tab-container"},H={key:0,class:"tab__header"},W={class:"tab__content-container"},O={class:"flex items-center with-warnings"},F=$(()=>d("span",null,"Warnings",-1)),G=h({__name:"TabsWidget",props:{tabs:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},isEmpty:{type:Boolean,required:!1,default:!1},hasError:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},hasBorder:{type:Boolean,required:!1,default:!1},initialTabOverride:{type:String,required:!1,default:null}},emits:["on-tab-change"],setup(e,{emit:b}){const o=e,i=y(o.initialTabOverride&&`#${o.initialTabOverride}`),_=v(()=>o.tabs.map(a=>a.hash.replace("#","")));function g(a){S.logger.info(w.TABS_TAB_CHANGE,{data:{newTab:a}}),b("on-tab-change",a)}return(a,r)=>(s(),n("div",x,[e.isLoading?(s(),c(N,{key:0})):e.error!==null?(s(),c(I,{key:1,error:e.error},null,8,["error"])):(s(),n(B,{key:2},[a.$slots.tabHeader?(s(),n("header",H,[u(a.$slots,"tabHeader",{},void 0,!0)])):T("",!0),f(),d("div",W,[p(l(V),{modelValue:i.value,"onUpdate:modelValue":r[0]||(r[0]=t=>i.value=t),tabs:e.tabs,onChanged:r[1]||(r[1]=t=>g(t))},k({"warnings-anchor":m(()=>[d("span",O,[p(l(q),{class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"16"}),f(),F])]),_:2},[E(l(_),t=>({name:t,fn:m(()=>[u(a.$slots,t,{},void 0,!0)])}))]),1032,["modelValue","tabs"])])],64))]))}});const D=C(G,[["__scopeId","data-v-36f2b969"]]);export{D as T};
