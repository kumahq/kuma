import{d as h,r as y,f as v,o as s,j as n,c,F as B,J as u,A as T,e as f,l as d,a as p,K as k,n as E,w as b,u as l,b as S,M as w,G as q,H as V,C,D as L,E as A}from"./index.60b0f0ac.js";import{E as N}from"./ErrorBlock.2ee4d08e.js";import{_ as $}from"./LoadingBlock.vue_vue_type_script_setup_true_lang.be7e4bb1.js";const x=e=>(C("data-v-6b9fc53d"),e=e(),L(),e),H={class:"tab-container","data-testid":"tab-container"},I={key:0,class:"tab__header"},W={class:"tab__content-container"},O={class:"flex items-center with-warnings"},F=x(()=>d("span",null,"Warnings",-1)),G=h({__name:"TabsWidget",props:{tabs:{type:Array,required:!0},isLoading:{type:Boolean,required:!1,default:!1},isEmpty:{type:Boolean,required:!1,default:!1},hasError:{type:Boolean,required:!1,default:!1},error:{type:[Error,null],required:!1,default:null},hasBorder:{type:Boolean,required:!1,default:!1},initialTabOverride:{type:String,required:!1,default:null}},emits:["on-tab-change"],setup(e,{emit:_}){const o=e,i=y(o.initialTabOverride&&`#${o.initialTabOverride}`),m=v(()=>o.tabs.map(a=>a.hash.replace("#","")));function g(a){q.logger.info(V.TABS_TAB_CHANGE,{data:{newTab:a}}),_("on-tab-change",a)}return(a,r)=>(s(),n("div",H,[e.isLoading?(s(),c($,{key:0})):e.error!==null?(s(),c(N,{key:1,error:e.error},null,8,["error"])):(s(),n(B,{key:2},[a.$slots.tabHeader?(s(),n("header",I,[u(a.$slots,"tabHeader",{},void 0,!0)])):T("",!0),f(),d("div",W,[p(l(w),{modelValue:i.value,"onUpdate:modelValue":r[0]||(r[0]=t=>i.value=t),tabs:e.tabs,onChanged:r[1]||(r[1]=t=>g(t))},k({"warnings-anchor":b(()=>[d("span",O,[p(l(S),{class:"mr-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"16"}),f(),F])]),_:2},[E(l(m),t=>({name:t,fn:b(()=>[u(a.$slots,t,{},void 0,!0)])}))]),1032,["modelValue","tabs"])])],64))]))}});const J=A(G,[["__scopeId","data-v-6b9fc53d"]]);export{J as T};
