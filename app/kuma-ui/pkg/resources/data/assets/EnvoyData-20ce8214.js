import{d as p,o as e,e as c,a as s,w as t,i,t as f,b as n,s as m,F as k,h as _,D as g,g as u}from"./index-f1b8ae6a.js";import{_ as v}from"./CodeBlock.vue_vue_type_style_index_0_lang-d1d1c408.js";import{h,E as q,n as E,o as b,i as x,f as D}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";const S={class:"envoy-data-actions"},B=p({__name:"EnvoyData",props:{status:{type:String,required:!0},resource:{type:String,required:!0},src:{type:String,required:!0},queryKey:{type:String,required:!0}},setup(y){const r=y,{t:l}=h();return(C,N)=>(e(),c("div",null,[r.status!=="online"?(e(),s(n(m),{key:0,appearance:"info"},{alertMessage:t(()=>[i("p",null,f(n(l)("common.detail.no_envoy_data",{resource:r.resource})),1)]),_:1})):(e(),s(x,{key:1,src:r.src},{default:t(({data:a,error:o,refresh:d})=>[o?(e(),s(q,{key:0,error:o},null,8,["error"])):a===void 0?(e(),s(E,{key:1})):a===""?(e(),s(b,{key:2})):(e(),c(k,{key:3},[i("div",S,[_(n(g),{appearance:"primary",icon:"redo","data-testid":"envoy-data-refresh-button",onClick:d},{default:t(()=>[u(`
            Refresh
          `)]),_:2},1032,["onClick"])]),u(),_(v,{id:"code-block-envoy-data",language:"json",code:typeof a=="string"?a:JSON.stringify(a,null,2),"is-searchable":"","query-key":r.queryKey},null,8,["code","query-key"])],64))]),_:1},8,["src"]))]))}});const F=D(B,[["__scopeId","data-v-c1f432b0"]]);export{F as E};
