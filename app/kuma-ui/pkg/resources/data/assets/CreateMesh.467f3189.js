import{_ as p,a5 as m,Y as u,c as _,w as s,r as t,o as g,a as n,f as e,b as o,t as b}from"./index.8c6a97c0.js";import{O as h}from"./OnboardingNavigation.4221abdc.js";import{O as f,a as v}from"./OnboardingPage.71f46d6d.js";const y={name:"CreateMesh",components:{OnboardingNavigation:h,OnboardingHeading:f,OnboardingPage:v},data(){return{productName:m,tableHeaders:[{label:"Name",key:"name"},{label:"Services",key:"servicesAmount"},{label:"DPPs",key:"dppsAmount"}],tableData:{total:1,data:[{name:"default",servicesAmount:0,dppsAmount:0}]}}},computed:{...u({multicluster:"config/getMulticlusterStatus"}),previousStep(){return this.multicluster?"onboarding-multi-zone":"onboarding-configuration-types"}}},O={class:"text-center mb-4"},x=e("i",null,"default",-1),N={class:"flex justify-center mt-10 mb-12 pb-12"},C={class:"w-full sm:w-3/5 lg:w-2/5 p-4"},P=e("p",{class:"text-center"}," This mesh is empty. Next, you add services and their data plane proxies. ",-1);function k(w,A,D,H,a,r){const i=t("OnboardingHeading"),c=t("KTable"),d=t("OnboardingNavigation"),l=t("OnboardingPage");return g(),_(l,null,{header:s(()=>[n(i,{title:"Create the mesh"})]),content:s(()=>[e("p",O,[o(" When you install, "+b(a.productName)+" creates a ",1),x,o(" mesh, but you can add as many meshes as you need. ")]),e("div",N,[e("div",C,[n(c,{fetcher:()=>a.tableData,headers:a.tableHeaders,"disable-pagination":"","is-small":""},null,8,["fetcher","headers"])])]),P]),navigation:s(()=>[n(d,{"next-step":"onboarding-add-services","previous-step":r.previousStep},null,8,["previous-step"])]),_:1})}const B=p(y,[["render",k]]);export{B as default};
