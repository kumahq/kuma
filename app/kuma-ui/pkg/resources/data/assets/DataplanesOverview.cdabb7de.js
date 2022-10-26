import{C as D,ck as O,i as u,co as y,cR as x,o as h,c as w,w as i,a as r,k as b,l as n,t as _,z as k,j as s}from"./index.563e1198.js";import{L as P}from"./LoadingBox.ac5f58ac.js";import{O as C}from"./OnboardingNavigation.ff859691.js";import{O as L,a as N}from"./OnboardingPage.59c97d73.js";const B={name:"DataplanesOverview",components:{OnboardingNavigation:C,OnboardingHeading:L,OnboardingPage:N,LoadingBox:P},metaInfo(){return{title:this.title}},data(){return{productName:O,tableHeaders:[{label:"Mesh",key:"mesh"},{label:"Name",key:"name"},{label:"Status",key:"status"}],tableData:{total:0,data:[]},timeout:null}},computed:{title(){return this.tableData.data.length?"Success":"Waiting for DPPs"},description(){return this.tableData.data.length?"The following data plane proxies (DPPs) are connected to the control plane:":null}},created(){this.getAllDataplanes()},beforeUnmount(){clearTimeout(this.timeout)},methods:{async getAllDataplanes(){let c=!1;const d=[];try{const g=(await u.getAllDataplanes({size:10})).items;for(let t=0;t<g.length;t++){const{name:a,mesh:o}=g[t],{status:l}=await u.getDataplaneOverviewFromMesh({mesh:o,name:a}).then(m=>y(m.dataplaneInsight));l===x&&(c=!0),d.push({status:l,name:a,mesh:o})}}catch(p){console.error(p)}this.tableData.data=d,this.tableData.total=this.tableData.data.length,c&&(this.timeout=setTimeout(()=>{this.getAllDataplanes()},1e3))}}},T={key:0,class:"justify-center flex my-4"},A={key:1},F={class:"flex justify-center mt-10 mb-16 pb-16"},H={class:"w-full sm:w-3/5 p-4"},I={class:"font-bold mb-4"};function S(c,d,p,g,t,a){const o=s("OnboardingHeading"),l=s("LoadingBox"),m=s("KTable"),f=s("OnboardingNavigation"),v=s("OnboardingPage");return h(),w(v,null,{header:i(()=>[r(o,{title:a.title,description:a.description},null,8,["title","description"])]),content:i(()=>[t.tableData.data.length?(h(),b("div",A,[n("div",F,[n("div",H,[n("p",I," Found "+_(t.tableData.data.length)+" DPPs: ",1),r(m,{class:"onboarding-dataplane-table",fetcher:()=>t.tableData,headers:t.tableHeaders,"disable-pagination":"","is-small":""},{status:i(({rowValue:e})=>[n("div",{class:k(["entity-status",{"is-offline":e.toLowerCase()==="offline"||e===!1,"is-online":e.toLowerCase()==="online","is-degraded":e.toLowerCase()==="partially degraded","is-not-available":e.toLowerCase()==="not available"}])},[n("span",null,_(e),1)],2)]),_:1},8,["fetcher","headers"])])])])):(h(),b("div",T,[r(l)]))]),navigation:i(()=>[r(f,{"next-step":"onboarding-completed","previous-step":"onboarding-add-services-code","should-allow-next":t.tableData.data.length>0},null,8,["should-allow-next"])]),_:1})}const M=D(B,[["render",S]]);export{M as default};
