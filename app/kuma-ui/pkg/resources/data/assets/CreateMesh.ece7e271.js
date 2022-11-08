import{D as p,L as m,cm as u,cp as b,o as g,c as _,w as s,a as n,l as e,b as o,t as h,i as t}from"./index.c585bc0e.js";import{O as f}from"./OnboardingNavigation.bd8697ea.js";import{O as v,a as y}from"./OnboardingPage.bb2e2005.js";const O={name:"CreateMesh",components:{OnboardingNavigation:f,OnboardingHeading:v,OnboardingPage:y,KTable:m},data(){return{productName:u,tableHeaders:[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],tableData:{total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]}}},computed:{...b({multicluster:"config/getMulticlusterStatus"}),previousStep(){return this.multicluster?"onboarding-multi-zone":"onboarding-configuration-types"}}},x={class:"text-center mb-4"},N=e("i",null,"default",-1),C={class:"flex justify-center mt-10 mb-12 pb-12"},D={class:"w-full sm:w-3/5 lg:w-2/5 p-4"},P=e("p",{class:"text-center"}," This mesh is empty. Next, you add services and their data plane proxies. ",-1);function T(k,w,A,H,a,r){const i=t("OnboardingHeading"),c=t("KTable"),l=t("OnboardingNavigation"),d=t("OnboardingPage");return g(),_(d,null,{header:s(()=>[n(i,{title:"Create the mesh"})]),content:s(()=>[e("p",x,[o(" When you install, "+h(a.productName)+" creates a ",1),N,o(" mesh, but you can add as many meshes as you need. ")]),e("div",C,[e("div",D,[n(c,{fetcher:()=>a.tableData,headers:a.tableHeaders,"disable-pagination":"","is-small":""},null,8,["fetcher","headers"])])]),P]),navigation:s(()=>[n(l,{"next-step":"onboarding-add-services","previous-step":r.previousStep},null,8,["previous-step"])]),_:1})}const K=p(O,[["render",T]]);export{K as default};
